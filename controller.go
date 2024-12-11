package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Controller struct {
	ID        string
	broker    *Queue
	pubsub    *PubSubSystem
	isQueue   bool
	nextFree  time.Time
	mutex     sync.Mutex
	lastState map[string]string
}

func (c *Controller) ProcessQueueRequests() {
	if !c.isQueue {
		return
	}

	if c.nextFree.IsZero() {
		c.nextFree = time.Now()
	}
	if c.lastState == nil {
		c.lastState = make(map[string]string)
	}

	for {
		c.mutex.Lock()
		msg, ok := c.broker.Dequeue()
		c.mutex.Unlock()

		if ok {
			payload := msg.Payload.(map[string]string)
			aircraftID := payload["aircraft_id"]
			actionType := payload["type"]

			c.mutex.Lock()
			lastAction, exists := c.lastState[aircraftID]
			if exists && ((lastAction == "landing" && actionType != "takeoff") || (lastAction == "takeoff" && actionType != "landing")) {
				logMessage("WARN", "QueueController",
					fmt.Sprintf("Skipping invalid %s request from Aircraft %s. Last action was %s.", actionType, aircraftID, lastAction))
				c.mutex.Unlock()
				continue
			}
			c.mutex.Unlock()

			if time.Now().Before(c.nextFree) {
				logMessage("INFO", "QueueController",
					fmt.Sprintf("Delaying %s request from Aircraft %s until %s.", actionType, aircraftID, c.nextFree.Format(time.RFC3339)))
				time.Sleep(time.Until(c.nextFree))
			}

			logMessage("INFO", "QueueController",
				fmt.Sprintf("Processing %s request from Aircraft %s at %s.", actionType, aircraftID, c.nextFree.Format(time.RFC3339)))

			success := rand.Float32() > 0.8
			if success {
				logMessage("INFO", "QueueController",
					fmt.Sprintf("Successfully processed %s request from Aircraft %s.", actionType, aircraftID))

				c.mutex.Lock()
				c.nextFree = c.nextFree.Add(15 * time.Minute)
				c.lastState[aircraftID] = actionType
				c.mutex.Unlock()
			} else {
				logMessage("ERROR", "QueueController",
					fmt.Sprintf("Failed to process %s request from Aircraft %s. Re-enqueuing.", actionType, aircraftID))

				c.mutex.Lock()
				c.broker.Enqueue(msg)
				c.mutex.Unlock()

				failureMessage := fmt.Sprintf("Aircraft %s failed to process %s request.", aircraftID, actionType)
				c.pubsub.Publish(Message{
					Topic: "updates",
					Payload: map[string]string{
						"update_type": "Failure",
						"details":     failureMessage,
						"aircraft_id": aircraftID,
						"timestamp":   time.Now().Format(time.RFC3339),
					},
				})

				logMessage("INFO", "QueueController",
					fmt.Sprintf("Published failure update: %s", failureMessage))
			}

			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (c *Controller) ProcessPubSubUpdates() {
	if c.isQueue {
		return
	}

	ch := c.pubsub.Subscribe("updates")
	for msg := range ch {
		logMessage("INFO", "PubSubController", fmt.Sprintf("Received update: %v", msg.Payload))
	}
}

func (c *Controller) PublishUpdate(updateType string, details string) {
	msg := Message{
		Topic: "updates",
		Payload: map[string]string{
			"update_type": updateType,
			"details":     details,
			"timestamp":   time.Now().Format(time.RFC3339),
		},
	}
	c.pubsub.Publish(msg)
	logMessage("INFO", "PubSubController", fmt.Sprintf("Published update: %s - %s.", updateType, details))
}
