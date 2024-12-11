package main

import (
	"fmt"
	"sync"
	"time"
)

type Aircraft struct {
	ID        string
	broker    *Queue
	pubsub    *PubSubSystem
	lastState string
	mutex     sync.Mutex
}

func (a *Aircraft) SendRequest() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var nextRequestType string
	if a.lastState == "landing" {
		nextRequestType = "takeoff"
	} else {
		nextRequestType = "landing"
	}

	a.lastState = nextRequestType
	msg := Message{
		Topic: "queue",
		Payload: map[string]string{
			"aircraft_id": a.ID,
			"type":        nextRequestType,
			"timestamp":   time.Now().Format(time.RFC3339),
		},
	}
	a.broker.Enqueue(msg)
	fmt.Printf("Aircraft %s enqueued %s request.\n", a.ID, nextRequestType)
}

func (a *Aircraft) ListenForUpdates() {
	ch := a.pubsub.Subscribe("updates")
	go func() {
		for msg := range ch {
			fmt.Printf("Aircraft %s received update: %v\n", a.ID, msg.Payload)
		}
	}()
}
