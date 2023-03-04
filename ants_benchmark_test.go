package jant_test

import (
	"github.com/caser789/jant"
	"sync"
	"testing"
)

const RunTimes = 10000000

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				demoFunc()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkPoolGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			jant.Push(func() {
				defer wg.Done()
				demoFunc()
			})
		}
		wg.Wait()
	}
}
