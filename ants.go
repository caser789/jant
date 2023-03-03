package jant

import (
	"math"
	"runtime"
)

const DEFAULT_POOL_SIZE = math.MaxInt32

var defaultPool = NewPool(DEFAULT_POOL_SIZE, runtime.GOMAXPROCS(-1))

func Push(task f) error {
	return defaultPool.Push(task)
}

func Size() int {
	return int(defaultPool.Size())
}

func Cap() int {
	return int(defaultPool.Cap())
}
