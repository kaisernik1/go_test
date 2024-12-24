package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	const maxErrors = 3
	errorCount := 0

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

	for {
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error sending request: %v\n", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= maxErrors {
				log.Fatal("Unable to fetch server statistics.")
			}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading body: %v\n", err)
		}

		values, err := parseCSV(string(body))
		if err != nil {
			log.Fatalf("Error parsing CSV: %v\n", err)
		}

		processValues(values)
		break
	}
}

func parseCSV(data string) ([]int64, error) {
	reader := csv.NewReader(strings.NewReader(data))
	record, err := reader.Read()
	if err != nil {
		return nil, err
	}

	var values []int64
	for _, field := range record {
		value, err := strconv.ParseInt(field, 10, 64)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}

func processValues(values []int64) {
	loadAverage := float64(values[0])
	memTotalBytes := int64(values[1])
	memUsedBytes := int64(values[2])
	diskTotalBytes := int64(values[3])
	diskUsedBytes := int64(values[4])
	netBandwidthBps := int64(values[5])
	netUsageBps := int64(values[6])

	if loadAverage > 30 {
		fmt.Printf("Load Average is too high: %.2f\n", loadAverage)
	}

	memUsagePercent := float64(memUsedBytes) / float64(memTotalBytes) * 100
	if memUsagePercent > 80 {
		fmt.Printf("Memory usage too high: %.2f%%\n", memUsagePercent)
	}

	diskUsagePercent := float64(diskUsedBytes) / float64(diskTotalBytes) * 100
	if diskUsagePercent > 90 {
		freeMbLeft := float64((diskTotalBytes - diskUsedBytes)) / 1048576
		fmt.Printf("Free disk space is too low: %.2f Mb left\n", freeMbLeft)
	}

	netUsagePercent := float64(netUsageBps) / float64(netBandwidthBps) * 100
	if netUsagePercent > 90 {
		freeMbitsPerSec := float64((netBandwidthBps - netUsageBps) / 125000)
		fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeMbitsPerSec)
	}
}
