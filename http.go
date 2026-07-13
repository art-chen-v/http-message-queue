package main

import (
	"context"
	"net/http"
	"strconv"
	"time"
)
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
