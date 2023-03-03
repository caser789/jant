package jant

import (
	"fmt"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	p := NewPool(2, 3)
	p.Push(func() {
		fmt.Printf("func 1")
	})
	p.Push(func() {
		fmt.Printf("func 2")
	})
	p.Push(func() {
		fmt.Printf("func 3")
	})
	time.Sleep(time.Second * 3)
	p.Destroy()
}
