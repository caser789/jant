package jant_test

import (
	"testing"

	"github.com/caser789/jant"
)

var size = 1000

func BenchmarkPoolGroutine(b *testing.B) {
	p := jant.NewPool(size, 32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Push(demoFunc)
	}
}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		go demoFunc()
	}
}