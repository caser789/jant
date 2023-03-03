package jant_test

import (
	"testing"

	"github.com/caser789/jant"
)

func BenchmarkPoolGroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jant.Push(demoFunc)
	}
}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		go demoFunc()
	}
}
