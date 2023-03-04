package jant

import (
	"math"
	"sync"
	"sync/atomic"
)

type pf func(interface{}) error

// PoolWithFunc accept the tasks from client,it will limit the total
// of goroutines to a given number by recycling goroutines.
type PoolWithFunc struct {
	// capacity of the pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// signal is used to notice pool there are available
	// workers which can be sent to work.
	freeSignal chan sig

	// workers is a slice that store the available workers.
	workers []*WorkerWithFunc

	// release is used to notice the pool to closed itself.
	release chan sig

	lock sync.Mutex

	once     sync.Once
	poolFunc pf
}

// NewPoolWithFunc generates a instance of ants pool with a specific function.
func NewPoolWithFunc(size int, f pf) (*PoolWithFunc, error) {
	if size <= 0 {
		return nil, ErrPoolSizeInvalid
	}

	p := &PoolWithFunc{
		capacity:   int32(size),
		freeSignal: make(chan sig, math.MaxInt32),
		release:    make(chan sig, 1),
		poolFunc:   f,
	}

	return p, nil
}

//-------------------------------------------------------------------------

// Push submit a task to pool
func (p *PoolWithFunc) Serve(args interface{}) error {
	if len(p.release) > 0 {
		return ErrPoolClosed
	}

	w := p.getWorker()
	w.sendTask(args)
	return nil
}

// Running returns the number of the currently running goroutines
func (p *PoolWithFunc) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns the available goroutines to work
func (p *PoolWithFunc) Free() int {
	return int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
}

// Cap returns the capacity of this pool
func (p *PoolWithFunc) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Release Closed this pool
func (p *PoolWithFunc) Release() error {
	p.once.Do(func() {
		p.release <- sig{}
		running := p.Running()
		for i := 0; i < running; i++ {
			p.getWorker().stop()
		}
	})
	return nil
}

// ReSize change the capacity of this pool
func (p *PoolWithFunc) ReSize(size int) {
	if size == p.Cap() {
		return
	}

	if size < p.Cap() {
		diff := p.Cap() - size
		for i := 0; i < diff; i++ {
			p.getWorker().stop()
		}
	}
	atomic.StoreInt32(&p.capacity, int32(size))
}

//-------------------------------------------------------------------------

// getWorker returns a available worker to run the tasks.
func (p *PoolWithFunc) getWorker() *WorkerWithFunc {
	var w *WorkerWithFunc
	waiting := false

	p.lock.Lock()
	workers := p.workers
	n := len(workers) - 1
	if n < 0 {
		if p.running >= p.capacity {
			waiting = true
		} else {
			p.running++
		}
	} else {
		w = workers[n]
		workers[n] = nil
		p.workers = workers[:n]
	}
	p.lock.Unlock()

	if waiting {
		<-p.freeSignal
		for {
			p.lock.Lock()
			workers = p.workers
			l := len(workers) - 1
			if l < 0 {
				p.lock.Unlock()
				continue
			}
			w = workers[l]
			workers[l] = nil
			p.workers = workers[:l]
			p.lock.Unlock()
			break
		}
	} else if w == nil {
		w = &WorkerWithFunc{
			pool: p,
			args: make(chan interface{}),
		}
		w.run()
	}
	return w
}

// putWorker puts a worker back into free pool, recycling the goroutines.
func (p *PoolWithFunc) putWorker(worker *WorkerWithFunc) {
	p.lock.Lock()
	p.workers = append(p.workers, worker)
	p.lock.Unlock()
	p.freeSignal <- sig{}
}
