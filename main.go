package main

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "log"
    "net/http"
    "io"
    "strconv"
    "strings"
    "time"
)

const (
    loadAverageThreshold   = 30.0
    memoryUsageThreshold   = 80.0
    freeDiskSpaceThreshold = 10.0
    networkBandwidthThresh = 90.0
    maxErrors              = 3
    pollInterval           = time.Second * 60
    serverURL              = "http://srv.msk01.gigacorp.local/_stats"
)

func main() {
    errorCount := 0
    for {
        stats, err := getServerStats()
        if err != nil {
            log.Println("Error fetching stats:", err)
            errorCount++
            if errorCount >= maxErrors {
                fmt.Println("Unable to fetch server statistics.")
                return
            }
            time.Sleep(pollInterval)
            continue
        }
        errorCount = 0

        loadAvg, memoryTotal, memoryUsed, diskTotal, diskUsed, networkBandwidth, networkUsage := parseStats(stats)

        checkLoadAverage(loadAvg)
        checkMemoryUsage(memoryTotal, memoryUsed)
        checkFreeDiskSpace(diskTotal, diskUsed)
        checkNetworkBandwidth(networkBandwidth, networkUsage)

        time.Sleep(pollInterval)
    }
}

func getServerStats() ([]byte, error) {
    resp, err := http.Get(serverURL)
    if err != nil {
        return nil, fmt.Errorf("failed to make HTTP request: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    return body, nil
}

func parseStats(data []byte) (float64, int64, int64, int64, int64, float64, float64) {
    reader := csv.NewReader(bytes.NewReader(data))
    fields, err := reader.Read()
    if err != nil {
        log.Fatalf("Failed to parse CSV data: %v", err)
    }

    loadAvgStr := strings.Trim(fields[0], ",")
    loadAvg, _ := strconv.ParseFloat(loadAvgStr, 64)

    memoryTotal, _ := strconv.ParseInt(fields[1], 10, 64)
    memoryUsed, _ := strconv.ParseInt(fields[2], 10, 64)

    diskTotal, _ := strconv.ParseInt(fields[3], 10, 64)
    diskUsed, _ := strconv.ParseInt(fields[4], 10, 64)

    networkBandwidth, _ := strconv.ParseFloat(fields[5], 64)
    networkUsage, _ := strconv.ParseFloat(fields[6], 64)

    return loadAvg, memoryTotal, memoryUsed, diskTotal, diskUsed, networkBandwidth, networkUsage
}

func checkLoadAverage(loadAvg float64) {
    if loadAvg > loadAverageThreshold {
        fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
    }
}

func checkMemoryUsage(memoryTotal, memoryUsed int64) {
    usagePercent := float64(memoryUsed) / float64(memoryTotal) * 100
    if usagePercent > memoryUsageThreshold {
        fmt.Printf("Memory usage too high: %.2f%%\n", usagePercent)
    }
}

func checkFreeDiskSpace(diskTotal, diskUsed int64) {
    freeSpace := float64(diskTotal-diskUsed) / 1024 / 1024
    if freeSpace < freeDiskSpaceThreshold {
        fmt.Printf("Free disk space is too low: %.2f MB left\n", freeSpace)
    }
}

func checkNetworkBandwidth(bandwidth, usage float64) {
    usagePercent := usage / bandwidth * 100
    if usagePercent > networkBandwidthThresh {
        fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", bandwidth*0.1)
    }
}