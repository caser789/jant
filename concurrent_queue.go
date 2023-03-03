package jant

import (
	"container/list"
	"sync"
)

type ConcurrentQueue struct {
	queue *list.List
	m     sync.Mutex
}

func NewConcurrentQueue() *ConcurrentQueue {
	q := new(ConcurrentQueue)
	q.queue = list.New()
	return q
}

func (q *ConcurrentQueue) push(v interface{}) {
	defer q.m.Unlock()
	q.m.Lock()
	q.queue.PushFront(v)
}

func (q *ConcurrentQueue) pop() interface{} {
	defer q.m.Unlock()
	q.m.Lock()
	if elem := q.queue.Back(); elem != nil {
		return q.queue.Remove(elem)
	}

	return nil
}
