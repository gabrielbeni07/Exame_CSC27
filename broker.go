package main

import (
	"fmt"
	"sync"
)

var (
	queueState   []Message
	updatesState []Message
)

type Message struct {
	Topic      string
	Payload    interface{}
	RetryCount int
}

type Queue struct {
	requests []Message
	mutex    sync.Mutex
}

type PubSubSystem struct {
	subscribers map[string][]chan Message
	mutex       sync.RWMutex
}

func (q *Queue) Enqueue(msg Message) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.requests = append(q.requests, msg)
	queueState = append(queueState, msg)
	logMessage("INFO", "Queue", fmt.Sprintf("Message added to queue: %+v", msg))
}

func (q *Queue) Dequeue() (Message, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.requests) == 0 {
		return Message{}, false
	}
	msg := q.requests[0]
	q.requests = q.requests[1:]
	return msg, true
}

func (ps *PubSubSystem) Subscribe(topic string) chan Message {
	ch := make(chan Message, 1)
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	if ps.subscribers == nil {
		ps.subscribers = make(map[string][]chan Message)
	}
	ps.subscribers[topic] = append(ps.subscribers[topic], ch)
	return ch
}

func (ps *PubSubSystem) Publish(msg Message) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	subs, found := ps.subscribers[msg.Topic]
	if found {
		for _, ch := range subs {
			ch <- msg
		}
	}
	updatesState = append(updatesState, msg)
	logMessage("INFO", "PubSub", fmt.Sprintf("Message published: %+v", msg))
}
