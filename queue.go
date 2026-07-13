package main

import (
	"container/list"
	"context"
	"sync"
)


type Queue struct {
	messages *list.List
	waiters  *list.List
	mu       sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		messages: list.New(),
		waiters:  list.New(),
	}
}

func (q *Queue) Enqueue(msg string) {
	q.mu.Lock()
	if q.waiters.Len() > 0 {
		front := q.waiters.Front()
		ch := front.Value.(chan string)
		q.waiters.Remove(front)
		q.mu.Unlock()
		ch <- msg
		return
	}
	q.messages.PushBack(msg)
	q.mu.Unlock()
}

func (q *Queue) Dequeue(ctx context.Context) (string, bool) {
	q.mu.Lock()
	if q.messages.Len() > 0 {
		el := q.messages.Front()
		q.messages.Remove(el)
		msg := el.Value.(string)
		q.mu.Unlock()
		return msg, true
	}

	ch := make(chan string, 1)
	el := q.waiters.PushBack(ch)
	q.mu.Unlock()

	select {
	case msg := <-ch:
		return msg, true
	case <-ctx.Done():
		q.mu.Lock()
		q.waiters.Remove(el)
		q.mu.Unlock()
		return "", false
	}
}
