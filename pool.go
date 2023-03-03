package jant

import (
	"math"
	"sync"
	"sync/atomic"
)

type sig struct{}
type f func()

type Pool struct {
	tasks        chan f
	length       int32
	capacity     int32
	workers      chan *Worker
	destroy      chan sig
	m            sync.Mutex
	processCount int
}

func NewPool(size int, processCount int) *Pool {
	p := &Pool{
		capacity:     int32(size),
		tasks:        make(chan f, math.MaxInt32),
		destroy:      make(chan sig, processCount),
		workers:      make(chan *Worker, size),
		processCount: processCount,
	}
	p.loop()
	return p
}

func (p *Pool) Push(task f) error {
	if len(p.destroy) > 0 {
		return nil
	}

	p.tasks <- task
	return nil
}

func (p *Pool) Size() int32 {
	return atomic.LoadInt32(&p.length)
}

func (p *Pool) Cap() int32 {
	return atomic.LoadInt32(&p.capacity)
}

func (p *Pool) Destroy() error {
	p.m.Lock()
	defer p.m.Unlock()

	for i := 0; i < p.processCount+1; i++ {
		p.destroy <- sig{}
	}
	return nil
}

func (p *Pool) loop() {
	for i := 0; i < p.processCount; i++ {
		go func() {
			for {
				select {
				case task := <-p.tasks:
					p.getWorker().sendTask(task)
				case <-p.destroy:
					return
				}
			}
		}()
	}
}

func (p *Pool) reachLimit() bool {
	return p.Size() >= p.Cap()
}

func (p *Pool) newWorker() *Worker {
	worker := &Worker{
		pool:  p,
		tasks: make(chan f),
		exit:  make(chan sig),
	}
	worker.Run()
	atomic.AddInt32(&p.length, 1)
	p.workers <- worker
	return worker
}

func (p *Pool) getWorker() *Worker {
	var worker *Worker
	if p.reachLimit() {
		worker = <-p.workers
		return worker
	}

	worker = p.newWorker()
	return worker
}
