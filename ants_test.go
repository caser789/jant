package jant_test

import (
	"runtime"
	"testing"

	"github.com/caser789/jant"
)

var n = 100000

func demoFunc() {
	for i := 0; i < 1000000; i++ {
	}
}

func TestDefaultPool(t *testing.T) {
	for i := 0; i < n; i++ {
		jant.Push(demoFunc)
	}

	t.Logf("pool capacity:%d", jant.Cap())
	t.Logf("running workers:%d", jant.Running())
	t.Logf("free workers:%d", jant.Free())

	// jant.Wait()

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d", mem.TotalAlloc/1024)
}

func TestNoPool(t *testing.T) {
	for i := 0; i < n; i++ {
		go demoFunc()
	}

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d", mem.TotalAlloc/1024)
}

func TestCustomPool(t *testing.T) {
	p := jant.NewPool(1000, 100)
	for i := 0; i < n; i++ {
		p.Push(demoFunc)
	}

	t.Logf("pool capacity:%d", p.Cap())
	t.Logf("running workers number:%d", p.Running())
	t.Logf("free workers number:%d", p.Free())

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
}
