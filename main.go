package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Stats struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := "http://srv.msk01.gigacorp.local/_stats"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v\n", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v\n", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading body: %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: Status code is %d\n", resp.StatusCode)
	}

	var stats Stats
	err = json.Unmarshal(body, &stats)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v\n", err)
	}

	fmt.Printf("CPU Usage: %.2f%%\n", stats.CPUUsage)
	fmt.Printf("Memory Usage: %.2f%%\n", stats.MemoryUsage)
	fmt.Printf("Disk Usage: %.2f%%\n", stats.DiskUsage)
}
