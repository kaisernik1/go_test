package main

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "io"
    "log"
    "math"
    "net/http"
    "strconv"
    "strings"
    "time"
)

const (
    loadAverageThreshold   = 30
    memoryUsageThreshold   = 80
    freeDiskSpaceThreshold = 10
    networkBandwidthThresh = 90
    maxErrors              = 3
    pollInterval           = time.Second * 5
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

func parseStats(data []byte) (int, int, int, int, int, int, int) {
    reader := csv.NewReader(bytes.NewReader(data))
    fields, err := reader.Read()
    if err != nil {
        log.Fatalf("Failed to parse CSV data: %v", err)
    }

    if len(fields) != 7 {
        log.Fatalf("Invalid number of fields in CSV file: expected 7, got %d", len(fields))
    }

    loadAvgStr := strings.Trim(fields[0], ",")
    loadAvg, err := strconv.Atoi(loadAvgStr)
    if err != nil {
        log.Fatalf("Failed to convert Load Average to integer: %v", err)
    }

    memoryTotal, err := strconv.Atoi(fields[1])
    if err != nil {
        log.Fatalf("Failed to convert Memory Total to integer: %v", err)
    }

    memoryUsed, err := strconv.Atoi(fields[2])
    if err != nil {
        log.Fatalf("Failed to convert Memory Used to integer: %v", err)
    }

    diskTotal, err := strconv.Atoi(fields[3])
    if err != nil {
        log.Fatalf("Failed to convert Disk Total to integer: %v", err)
    }

    diskUsed, err := strconv.Atoi(fields[4])
    if err != nil {
        log.Fatalf("Failed to convert Disk Used to integer: %v", err)
    }

    networkBandwidth, err := strconv.Atoi(fields[5])
    if err != nil {
        log.Fatalf("Failed to convert Network Bandwidth to integer: %v", err)
    }

    networkUsage, err := strconv.Atoi(fields[6])
    if err != nil {
        log.Fatalf("Failed to convert Network Usage to integer: %v", err)
    }

    return loadAvg, memoryTotal, memoryUsed, diskTotal, diskUsed, networkBandwidth, networkUsage
}

func checkLoadAverage(loadAvg int) {
    if loadAvg > loadAverageThreshold {
        fmt.Printf("Load Average is too high: %d\n", loadAvg)
    }
}

func checkMemoryUsage(memoryTotal, memoryUsed int) {
    if memoryTotal == 0 {
        log.Fatalf("Memory Total cannot be zero")
    }

    usagePercent := int(math.Round(float64(memoryUsed) / float64(memoryTotal) * 100))
    if usagePercent > memoryUsageThreshold {
        fmt.Printf("Memory usage too high: %d%%\n", usagePercent)
    }
}

func checkFreeDiskSpace(diskTotal, diskUsed int) {
    freeSpace := int(math.Round(float64(diskTotal-diskUsed) / 1024 / 1024))
    if freeSpace < freeDiskSpaceThreshold {
        fmt.Printf("Free disk space is too low: %d Mb left\n", freeSpace)
    }
}

// Обновленная функция для перевода значений в мегабиты
func checkNetworkBandwidth(bandwidth, usage int) {
    if bandwidth == 0 {
        log.Fatalf("Network Bandwidth cannot be zero")
    }

    bandwidthMbps := bandwidth / 125000 // Переводим из байт в мегабиты
    usageMbps := usage / 125000         // Переводим из байт в мегабиты

    usagePercent := int(math.Round(float64(usageMbps) / float64(bandwidthMbps) * 100))
    if usagePercent > networkBandwidthThresh {
        fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", bandwidthMbps)
    }
}