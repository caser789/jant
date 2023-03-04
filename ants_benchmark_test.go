package jant_test

import (
	"github.com/caser789/jant"
	"sync"
	"testing"
	"time"
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
	GiB // 1073741824
	TiB // 1099511627776             (超过了int32的范围)
	PiB // 1125899906842624
	EiB // 1152921504606846976
	ZiB // 1180591620717411303424    (超过了int64的范围)
	YiB // 1208925819614629174706176
)
const RunTimes = 10000000
const loop = 10

func demoFunc() error {
	n := 10
	time.Sleep(time.Duration(n) * time.Millisecond)
	return nil
}

func demoPoolFunc(args interface{}) error {
	n := args.(int)
	time.Sleep(time.Duration(n) * time.Millisecond)
	return nil
}

func BenchmarkGoroutineWithFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			go func() {
				demoPoolFunc(loop)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkAntsPoolWithFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		p, _ := jant.NewPoolWithFunc(50000, func(i interface{}) error {
			demoPoolFunc(i)
			wg.Done()
			return nil
		})
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			p.Serve(loop)
		}
		wg.Wait()
		// b.Logf("running goroutines: %d", p.Running())
	}
}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			go demoPoolFunc(loop)
		}
	}
}

func BenchmarkPoolGoroutine(b *testing.B) {
	p, _ := jant.NewPoolWithFunc(50000, demoPoolFunc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			p.Serve(loop)
		}
	}
}
