package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func logMessage(level, context, message string) {
	logEntry := map[string]string{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level,
		"context":   context,
		"message":   message,
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Printf("[%s] [%s] %s: %s\n", level, context, time.Now().Format(time.RFC3339), message)
		return
	}

	fmt.Println(string(jsonData))
}
