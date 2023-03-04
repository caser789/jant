package jant

import "sync/atomic"

type Worker struct {
	pool  *Pool
	tasks chan f
}

func (w *Worker) Run() {
	go func() {
		for ff := range w.tasks {
			if ff == nil {
				atomic.AddInt32(&w.pool.running, -1)
				return
			}
			ff()
			w.pool.putWorker(w)
		}
	}()
}

func (w *Worker) Stop() {
	w.tasks <- nil
}

func (w *Worker) sendTask(task f) {
	w.tasks <- task
}
