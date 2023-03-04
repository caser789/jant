package jant

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
)

type sig struct{}
type f func()

// Pool accept the tasks from client,it will limit the total
// of goroutines to a given number by recycling goroutines.
type Pool struct {
	// capacity of the pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// signal is used to notice pool there are available
	// workers which can be sent to work.
	freeSignal chan sig

	// workers is a slice that store the available workers.
	workers []*Worker

	// workerPool is a pool that saves a set of temporary objects.
	workerPool sync.Pool

	// release is used to notice the pool to closed itself.
	release chan sig
	// closed is used to confirm whether this pool has been closed.
	closed int32

	lock         sync.Mutex
	processCount int
}

// Errors for the Ants API
var (
	ErrPoolSizeInvalid = errors.New("invalid size for pool")
	ErrPoolClosed      = errors.New("this pool has been closed")
)

func NewPool(size int, processCount int) (*Pool, error) {
	if size <= 0 || processCount <= 0 {
		return nil, ErrPoolSizeInvalid
	}

	p := &Pool{
		capacity:     int32(size),
		processCount: processCount,
		freeSignal:   make(chan sig, math.MaxInt32),
		release:      make(chan sig),
		closed:       0,
	}
	return p, nil
}

func (p *Pool) Push(task f) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	w := p.getWorker()
	w.sendTask(task)
	return nil
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Free() int {
	return int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
}

func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

func (p *Pool) Release() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	atomic.StoreInt32(&p.closed, 1)
	close(p.release)
	return nil
}

// ReSize change the capacity of this pool
func (p *Pool) ReSize(size int) {
	atomic.StoreInt32(&p.capacity, int32(size))
}

func (p *Pool) getWorker() *Worker {
	var w *Worker
	waiting := false
	p.lock.Lock()
	workers := p.workers
	n := len(workers) - 1
	if n < 0 {
		if p.running >= p.capacity {
			waiting = true
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
			n = len(workers) - 1
			if n < 0 {
				p.lock.Unlock()
				continue
			}
			w = workers[n]
			workers[n] = nil
			p.workers = workers[:n]
			p.lock.Unlock()
			break
		}
	} else {
		wp := p.workerPool.Get()
		if wp == nil {
			w = &Worker{
				pool: p,
			}
			w.Run()
			atomic.AddInt32(&p.running, 1)
		} else {
			w = wp.(*Worker)
		}
	}
	return w
}

func (p *Pool) putWorker(w *Worker) {
	p.workerPool.Put(w)
	p.lock.Lock()
	p.workers = append(p.workers, w)
	p.lock.Unlock()
	p.freeSignal <- sig{}
}
