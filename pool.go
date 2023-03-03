package jant

import (
	"math"
	"sync"
	"sync/atomic"
)

type sig struct{}
type f func()

type Pool struct {
	capacity     int32
	running      int32
	tasks        *ConcurrentQueue
	workers      *ConcurrentQueue
	freeSignal   chan sig
	launchSignal chan sig
	destroy      chan sig
	processCount int
	m            *sync.Mutex
	// wg           *sync.WaitGroup
}

func NewPool(size int, processCount int) *Pool {
	p := &Pool{
		capacity:     int32(size),
		processCount: processCount,
		tasks:        NewConcurrentQueue(),
		workers:      NewConcurrentQueue(),
		freeSignal:   make(chan sig, math.MaxInt32),
		launchSignal: make(chan sig, math.MaxInt32),
		destroy:      make(chan sig, processCount),
		// wg:           &sync.WaitGroup{},
		m: &sync.Mutex{},
	}
	p.loop()
	return p
}

func (p *Pool) Push(task f) error {
	if len(p.destroy) > 0 {
		return nil
	}

	p.tasks.push(task)
	p.launchSignal <- sig{}
	// p.wg.Add(1)
	return nil
}

// func (p *Pool) Wait() {
// 	p.wg.Wait()
// }

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Free() int {
	return int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
}

func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
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
				case <-p.launchSignal:
					p.getWorker().sendTask(p.tasks.pop().(f))
				case <-p.destroy:
					return
				}
			}
		}()
	}
}

func (p *Pool) reachLimit() bool {
	return p.Running() >= p.Cap()
}

func (p *Pool) newWorker() *Worker {
	if p.reachLimit() {
		<-p.freeSignal
		return p.getWorker()
	}

	worker := &Worker{
		pool:  p,
		tasks: make(chan f),
		exit:  make(chan sig),
	}
	worker.Run()
	return worker
}

func (p *Pool) getWorker() *Worker {
	defer atomic.AddInt32(&p.running, 1)
	if w := p.workers.pop(); w != nil {
		return w.(*Worker)
	}

	return p.newWorker()
}

func (p *Pool) PutWorker(w *Worker) {
	p.workers.push(w)
	if p.reachLimit() {
		p.freeSignal <- sig{}
	}
}
