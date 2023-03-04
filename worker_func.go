package jant

import (
	"sync/atomic"
)

// Worker is the actual executor who run the tasks,
// it will start a goroutine that accept tasks and
// perform function calls.
type WorkerWithFunc struct {
	// pool who owns this worker.
	pool *PoolWithFunc

	// args is a job should be done.
	args chan interface{}
}

// run will start a goroutine to repeat the process
// that perform the function calls.
func (w *WorkerWithFunc) run() {
	go func() {
		for args := range w.args {
			if args == nil || len(w.pool.release) > 0 {
				atomic.AddInt32(&w.pool.running, -1)
				return
			}

			w.pool.poolFunc(args)
			w.pool.putWorker(w)
		}
	}()
}

// stop this worker.
func (w *WorkerWithFunc) stop() {
	w.args <- nil
}

// sendTask send a task to this worker.
func (w *WorkerWithFunc) sendTask(args interface{}) {
	w.args <- args
}
