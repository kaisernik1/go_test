package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "net/http"
    "strconv"
    "strings"
    "sync"
    "time"
)

const (
    serverURL      = "http://srv.msk01.gigacorp.local/_stats"
    maxErrorsCount = 3 // Максимальное количество ошибок перед выдачей сообщения о невозможности сбора статистики
    pollInterval   = 60 * time.Second // Интервал опроса сервера (в секундах)
)

type ResourceStats struct {
    LoadAverage        int     `json:"load_average"`
    TotalMemory        float64 `json:"total_memory"`
    UsedMemory         float64 `json:"used_memory"`
    TotalDisk          float64 `json:"total_disk"`
    UsedDisk           float64 `json:"used_disk"`
    NetworkBandwidth   float64 `json:"network_bandwidth"`
    CurrentNetworkRate float64 `json:"current_network_rate"`
}

func main() {
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        for {
            errsCount := 0
            for errsCount < maxErrorsCount {
                stats, err := getResourceStats(serverURL)
                if err != nil {
                    log.Printf("Unable to fetch server statistics: %v", err)
                    errsCount++
                    continue
                }

                printWarnings(*stats)
                time.Sleep(pollInterval)
            }

            if errsCount >= maxErrorsCount {
                fmt.Fprintln(io.Discard, "Unable to fetch server statistics.")
            }
        }
    }()

    wg.Wait()
}

func printWarnings(stats ResourceStats) {
    if stats.LoadAverage > 30 {
        fmt.Printf("Load Average is too high: %d\n", stats.LoadAverage)
    }

    memUsagePercent := int(stats.UsedMemory / stats.TotalMemory * 100)
    if memUsagePercent > 80 {
        fmt.Printf("Memory usage too high: %d%%\n", memUsagePercent)
    }

    diskFreeMb := int((stats.TotalDisk - stats.UsedDisk) / 1024 / 1024)
    if diskFreeMb < 10 {
        fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeMb)
    }

    netAvailMbs := int((stats.NetworkBandwidth - stats.CurrentNetworkRate) / 125000)
    if netAvailMbs < int(0.1*stats.NetworkBandwidth/125000) {
        fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", netAvailMbs)
    }
}

func getResourceStats(url string) (*ResourceStats, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("unable to create request: %w", err)
    }

    req.Host = "srv.msk01.gigacorp.local"

    client := &http.Client{
        Timeout: 15 * time.Second,
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("unable to perform request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("unable to read response body: %w", err)
    }

    stats, err := parseResponseBody(body)
    if err != nil {
        return nil, fmt.Errorf("unable to parse response body: %w", err)
    }

    return stats, nil
}

func parseResponseBody(body []byte) (*ResourceStats, error) {
    values := strings.Split(string(body), ",")
    if len(values) != 7 {
        return nil, fmt.Errorf("incorrect number of values in response")
    }

    stats := &ResourceStats{}

    var err error
    stats.LoadAverage, err = strconv.Atoi(values[0])
    if err != nil {
        return nil, fmt.Errorf("unable to parse load average: %w", err)
    }

    stats.TotalMemory, err = strconv.ParseFloat(values[1], 64)
    if err != nil {
        return nil, fmt.Errorf("unable to parse total memory: %w", err)
    }

    stats.UsedMemory, err = strconv.ParseFloat(values[2], 64)
    if err != nil {
        return nil, fmt.Errorf("unable to parse used memory: %w", err)
    }

    stats.TotalDisk, err = strconv.ParseFloat(values[3], 64)
    if err != nil {
        return nil, fmt.Errorf("unable to parse total disk: %w", err)
    }

    stats.UsedDisk, err = strconv.ParseFloat(values[4], 64)
    if err != nil {
        return nil, fmt.Errorf("unable to parse used disk: %w", err)
    }

    stats.NetworkBandwidth, err = strconv.ParseFloat(values[5], 64)
    if err != nil {
        return nil, fmt.Errorf("unable to parse network bandwidth: %w", err)
    }

    stats.CurrentNetworkRate, err = strconv.ParseFloat(values[6], 64)
    if err != nil {
        return nil, fmt.Errorf("unable to parse current network rate: %w", err)
    }

    return stats, nil
}