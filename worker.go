package jant

import "sync/atomic"

type Worker struct {
	pool  *Pool
	exit  chan sig
	tasks chan f
}

func (w *Worker) Run() {
	go func() {
		for {
			select {
			case x := <-w.tasks:
				x()
				w.pool.workers.push(w)
				// w.pool.wg.Done()
			case <-w.exit:
				atomic.AddInt32(&w.pool.running, -1)
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.exit <- sig{}
}

func (w *Worker) sendTask(task f) {
	w.tasks <- task
}
