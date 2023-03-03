package jant

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPush(t *testing.T) {
	Push(func() {})
	Push(func() {})
	// assert.Equal(t, len(defaultPool.tasks), 2)
}

func TestSize(t *testing.T) {
	assert.Equal(t, int(defaultPool.Running()), Size())
}

func TestCapacity(t *testing.T) {
	assert.Equal(t, int(defaultPool.Cap()), Cap())
}
