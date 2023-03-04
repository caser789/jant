package jant_test

import (
	"runtime"
	"sync"
	"testing"

	"github.com/caser789/jant"
)

var n = 10000000

func demoFunc() {
	var m int
	for i := 0; i < 1000000; i++ {
		m += i
	}
}

func TestDefaultPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		jant.Push(func() {
			demoFunc()
			wg.Done()
		})
	}
	wg.Wait()

	t.Logf("running workers number:%d", jant.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d", mem.TotalAlloc/1024)
}

func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			demoFunc()
		}()
	}
	wg.Wait()

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d", mem.TotalAlloc/1024)
}
