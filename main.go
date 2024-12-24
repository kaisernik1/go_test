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

    loadAvgStr := strings.Trim(fields[0], ",")
    loadAvg, _ := strconv.Atoi(loadAvgStr)

    memoryTotal, _ := strconv.Atoi(fields[1])
    memoryUsed, _ := strconv.Atoi(fields[2])

    diskTotal, _ := strconv.Atoi(fields[3])
    diskUsed, _ := strconv.Atoi(fields[4])

    networkBandwidth, _ := strconv.Atoi(fields[5])
    networkUsage, _ := strconv.Atoi(fields[6])

    return loadAvg, memoryTotal, memoryUsed, diskTotal, diskUsed, networkBandwidth, networkUsage
}

func checkLoadAverage(loadAvg int) {
    if loadAvg > loadAverageThreshold {
        fmt.Printf("Load Average is too high: %d\n", loadAvg)
    }
}

func checkMemoryUsage(memoryTotal, memoryUsed int) {
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
    bandwidthMbps := bandwidth / 125000 // Переводим из байт в мегабиты
    usageMbps := usage / 125000         // Переводим из байт в мегабиты

    usagePercent := int(math.Round(float64(usageMbps) / float64(bandwidthMbps) * 100))
    if usagePercent > networkBandwidthThresh {
        fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", bandwidthMbps)
    }
}