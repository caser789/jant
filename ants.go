package jant

import (
	"runtime"
)

const DEFAULT_POOL_SIZE = 50000

var defaultPool, _ = NewPool(DEFAULT_POOL_SIZE, runtime.GOMAXPROCS(-1))

func Submit(task f) error {
	return defaultPool.Submit(task)
}

func Running() int {
	return defaultPool.Running()
}

func Cap() int {
	return defaultPool.Cap()
}

func Free() int {
	return defaultPool.Free()
}

func Release() {
	defaultPool.Release()
}
