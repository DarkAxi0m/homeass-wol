package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

func IsServerUpHTTP(url string, timeout time.Duration) bool {
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		log.Println("HTTP request error:", err)
		return false
	}
	defer resp.Body.Close()
	// Optionally check status code if expected
	return resp.StatusCode == http.StatusOK
}

func IsServerUpPingCli(host string, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", host)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("ping execution error: %s", strings.TrimSpace(string(output)))
	}
	return true, nil
}

func IsServerUpPing(host string, timeout time.Duration) bool {
	log.Println("Ping", host)
	pinger, err := probing.NewPinger(host)
	if err != nil {
		log.Println("Ping creation error:", err)
		return false
	}
	pinger.SetPrivileged(true)
	pinger.Timeout = timeout
	pinger.Count = 1

	err = pinger.Run() // Blocks until finished.
	if err != nil {
		log.Println("Ping execution error:", err)
		return false
	}

	stats := pinger.Statistics()
	log.Println(stats)
	log.Println(stats.PacketsRecv)
	return stats.PacketsRecv > 0
}
