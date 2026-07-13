package main

import "sync"

type Broker struct {
	queues map[string]*Queue
	mu     sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		queues: make(map[string]*Queue),
	}
}

func (b *Broker) getOrCreateQueue(name string) *Queue {
	b.mu.Lock()
	defer b.mu.Unlock()

	queue, ok := b.queues[name]
	if !ok {
		queue = NewQueue()
		b.queues[name] = queue
	}

	return queue
}

func (b *Broker) getQueue(name string) *Queue {
	b.mu.Lock()
	defer b.mu.Unlock()
	q, ok := b.queues[name]
	if !ok {
		return nil
	}
	return q
}
