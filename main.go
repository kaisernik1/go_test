package main


import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)


const (
	serverAddress = "srv.msk01.gigacorp.local"
	statsEndpoint = "/_stats"
	checkInterval = 10 * time.Second
	loadAverageThreshold = 30.0
	memoryUsageThreshold = 0.8
	diskUsageThreshold = 0.9
	networkBandwidthThreshold = 0.9
)


func fetchStats(client *http.Client) ([]string, error) {
	resp, err := client.Get("http://" + serverAddress + statsEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stats := strings.Split(string(bodyBytes), ",")
	if len(stats) != 7 {
		return nil, fmt.Errorf("unexpected number of stats: %d", len(stats))
	}
	return stats, nil
}

func checkResourceUsage(stats []string) {
	loadAvg, _ := strconv.ParseFloat(stats[0], 64)
	memTotal, _ := strconv.ParseInt(stats[1], 10, 64)
	memUsed, _ := strconv.ParseInt(stats[2], 10, 64)
	diskTotal, _ := strconv.ParseInt(stats[3], 10, 64)
	diskUsed, _ := strconv.ParseInt(stats[4], 10, 64)
	netTotal, _ := strconv.ParseInt(stats[5], 10, 64)
	netUsed, _ := strconv.ParseInt(stats[6], 10, 64)

	if loadAvg > loadAverageThreshold {
		fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
	}

	memUsage := float64(memUsed) / float64(memTotal)
	if memUsage > memoryUsageThreshold {
		fmt.Printf("Memory usage too high: %.0f%%\n", memUsage*100)
	}

	diskFree := (diskTotal - diskUsed) / (1024 * 1024)
	diskUsage := float64(diskUsed) / float64(diskTotal)
	if diskUsage > diskUsageThreshold {
		fmt.Printf("Free disk space is too low: %d Mb left\n", diskFree)
	}

	netUsage := float64(netUsed) / float64(netTotal)
	if netUsage > networkBandwidthThreshold {
		netFree := (netTotal - netUsed) / (1024 * 1024)
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", netFree)
	}
}

func main() {
	client := &http.Client{}
	errorCount := 0

	for {
		stats, err := fetchStats(client)
		if err != nil {
			errorCount++
			fmt.Fprintf(os.Stderr, "Error fetching stats: %v\n", err)
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				os.Exit(1)
			}
		} else {
			errorCount = 0
			checkResourceUsage(stats)
		}
		time.Sleep(checkInterval)
	}
}
