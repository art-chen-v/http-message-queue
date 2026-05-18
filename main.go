package main

import (
	"container/list"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Queue struct {
	messages *list.List
	waiters  *list.List
	mu       sync.Mutex
}

type Broker struct {
	queues map[string]*Queue
	mu     sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		messages: list.New(),
		waiters:  list.New(),
	}
}

func NewBroker() *Broker {
	return &Broker{
		queues: make(map[string]*Queue),
	}
}

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	broker := NewBroker()

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /{queue}", broker.addMsgToQueue)
	mux.HandleFunc("GET /{queue}", broker.getMsgFromQueue)

	server := &http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%d", *port),
	}

	log.Fatal(server.ListenAndServe())
}

func (b *Broker) addMsgToQueue(w http.ResponseWriter, r *http.Request) {
	queueName := r.PathValue("queue")
	message := r.URL.Query().Get("v")
	if message == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	queue := b.getOrCreateQueue(queueName)
	queue.Enqueue(message)
	w.WriteHeader(http.StatusOK)
}

func (b *Broker) getMsgFromQueue(w http.ResponseWriter, r *http.Request) {
	queueName := r.PathValue("queue")
	queue := b.getQueue(queueName)

	if queue == nil {
		http.Error(w, "сначала создайте очередь чере метод PUT", http.StatusNotFound)
		return
	}

	ctx := r.Context()

	timeoutStr := r.URL.Query().Get("timeout")
	if timeoutStr != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			http.Error(w, "неправильный формат timeout", http.StatusBadRequest)
			return
		}

		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
		defer cancel()
	}

	msg, ok := queue.Dequeue(ctx)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write([]byte(msg))
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
