package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var (
	queueLength         int
	successCount        int
	failureCount        int
	totalProcessingTime time.Duration
	processedCount      int
	runwayStatus        = "Open"
)

func main() {
	queue := &Queue{}
	pubsub := &PubSubSystem{}

	queueController := &Controller{
		ID:       "QueueController",
		broker:   queue,
		pubsub:   pubsub,
		isQueue:  true,
		nextFree: time.Now(),
	}

	pubSubController := &Controller{
		ID:      "PubSubController",
		broker:  queue,
		pubsub:  pubsub,
		isQueue: false,
	}

	go queueController.ProcessQueueRequests()

	go pubSubController.ProcessPubSubUpdates()

	go func() {
		updates := []struct {
			updateType string
			details    string
		}{
			{"Weather", "Heavy Rain"},
			{"Runway", "Runway 2 Closed"},
		}
		for {
			update := updates[rand.Intn(len(updates))]
			pubSubController.PublishUpdate(update.updateType, update.details)
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		aircrafts := []*Aircraft{
			{ID: "A1", broker: queue, pubsub: pubsub},
			{ID: "A2", broker: queue, pubsub: pubsub},
			{ID: "A3", broker: queue, pubsub: pubsub},
		}

		for _, aircraft := range aircrafts {
			aircraft.ListenForUpdates()
		}

		for {
			for _, aircraft := range aircrafts {
				aircraft.SendRequest()
				time.Sleep(2 * time.Second)
			}
		}
	}()

	http.HandleFunc("/dashboard", dashboardHandler)
	fmt.Println("Dashboard available at http://localhost:8080/dashboard")
	http.ListenAndServe(":8080", nil)
}
