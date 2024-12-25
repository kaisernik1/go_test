package main


import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"math"
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
	loadAvg, err := strconv.ParseFloat(stats[0], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing load average: %v\n", err)
		return
	}
	memTotal, err := strconv.ParseInt(stats[1], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing memTotal: %v\n", err)
		return
	}
	memUsed, err := strconv.ParseInt(stats[2], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing memUsed: %v\n", err)
		return
	}
	diskTotal, err := strconv.ParseInt(stats[3], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing diskTotal: %v\n", err)
		return
	}
	diskUsed, err := strconv.ParseInt(stats[4], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing diskUsed: %v\n", err)
		return
	}
	netTotal, err := strconv.ParseInt(stats[5], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing netTotal: %v\n", err)
		return
	}
	netUsed, err := strconv.ParseInt(stats[6], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing netUsed: %v\n", err)
		return
	}

	if loadAvg > loadAverageThreshold {
		fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
	}

    memUsage := float64(memUsed) / float64(memTotal)
	if memUsage >= memoryUsageThreshold { 
		fmt.Printf("Memory usage too high: %.0f%%\n", math.Round(memUsage*100)) 
	}

	diskFree := (diskTotal - diskUsed) / (1024 * 1024)
	diskUsage := float64(diskUsed) / float64(diskTotal)
	if diskUsage >= diskUsageThreshold { 
		fmt.Printf("Free disk space is too low: %d Mb left\n", diskFree)
	}

	netFree := ((netTotal / (1000 * 1000)) - (netUsed / (1000 * 1000)))
	netUsage := float64(netUsed) / float64(netTotal)
	if netUsage >= networkBandwidthThreshold { 
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
