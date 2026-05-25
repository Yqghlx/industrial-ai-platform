package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// WebSocket server address
	addr := os.Getenv("WS_ADDR")
	if addr == "" {
		addr = "localhost:8080"
	}

	// Number of concurrent connections
	numConns := 20
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &numConns)
	}

	// Test duration
	duration := 30 * time.Second

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	log.Printf("Connecting to %s with %d connections", u.String(), numConns)

	var wg sync.WaitGroup
	var successCount, failCount int64
	var mu sync.Mutex

	// Start connections
	for i := 0; i < numConns; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
				log.Printf("Connection %d failed: %v", id, err)
				return
			}
			defer conn.Close()

			mu.Lock()
			successCount++
			mu.Unlock()
			log.Printf("Connection %d established", id)

			// Send periodic messages for test duration
			start := time.Now()
			msgCount := 0

			for time.Since(start) < duration {
				// Send telemetry message
				msg := fmt.Sprintf(`{"type":"telemetry","device_id":"ws-test-%d","metrics":{"temperature":85.0,"vibration":3.5}}`, id)
				err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Printf("Connection %d write error: %v", id, err)
					return
				}
				msgCount++

				// Read response
				_, _, err = conn.ReadMessage()
				if err != nil {
					log.Printf("Connection %d read error: %v", id, err)
					return
				}

				time.Sleep(1 * time.Second)
			}

			log.Printf("Connection %d sent %d messages", id, msgCount)
		}(i)
	}

	wg.Wait()

	// Report
	fmt.Printf("\n=== WebSocket Stress Test Results ===\n")
	fmt.Printf("Target: %s\n", u.String())
	fmt.Printf("Connections: %d requested, %d success, %d failed\n", numConns, successCount, failCount)
	fmt.Printf("Duration: %s\n", duration)
	fmt.Printf("Success Rate: %.2f%%\n", float64(successCount)/float64(numConns)*100)
}