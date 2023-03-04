package jant

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
)

type sig struct{}
type f func() error

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

	// release is used to notice the pool to closed itself.
	release chan sig

	lock         sync.Mutex
	once         sync.Once
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
		release:      make(chan sig, 1),
	}
	return p, nil
}

func (p *Pool) Submit(task f) error {
	if len(p.release) > 0 {
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
	p.once.Do(func() {
		p.release <- sig{}
		running := p.Running()
		for i := 0; i < running; i++ {
			p.getWorker().Stop()
		}
	})
	return nil
}

// ReSize change the capacity of this pool
func (p *Pool) ReSize(size int) {
	if size == p.Cap() {
		return
	}

	if size < p.Cap() {
		diff := p.Cap() - size
		for i := 0; i < diff; i++ {
			p.getWorker().Stop()
		}
	}
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
		} else {
			p.running++
		}
	} else {
		<-p.freeSignal
		w = workers[n]
		workers[n] = nil
		p.workers = workers[:n]
	}
	p.lock.Unlock()

	if waiting {
		<-p.freeSignal
		p.lock.Lock()
		workers = p.workers
		n = len(workers) - 1
		w = workers[n]
		workers[n] = nil
		p.workers = workers[:n]
		p.lock.Unlock()
	} else if w == nil {
		w = &Worker{
			pool:  p,
			tasks: make(chan f),
		}
		w.Run()
	}
	return w
}

func (p *Pool) putWorker(w *Worker) {
	p.lock.Lock()
	p.workers = append(p.workers, w)
	p.lock.Unlock()
	p.freeSignal <- sig{}
}
