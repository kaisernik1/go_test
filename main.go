package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Stats struct {
	CpuUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
}

func main() {
	url := "http://srv.msk01.gigacorp.local/_stats"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Status code is %d\n", resp.StatusCode)
		return
	}

	var stats Stats
	err = json.Unmarshal(body, &stats)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return
	}

	fmt.Printf("CPU Usage: %.2f%%\n", stats.CpuUsage)
	fmt.Printf("Memory Usage: %.2f%%\n", stats.MemoryUsage)
	fmt.Printf("Disk Usage: %.2f%%\n", stats.DiskUsage)
}
