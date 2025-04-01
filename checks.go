package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/bougou/go-ipmi"
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

func PowerIpmi(host string, username string, password string, state string) bool {
	port := 623
	client, err := ipmi.NewClient(host, port, username, password)
	if err != nil {
		log.Fatalf("Error connecting to BMC: %v", err)
	}

	ctx := context.Background()

	// Connect will create an authenticated session for you.
	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	cont := ipmi.ChassisControlPowerUp
	if state == "OFF" {
		cont = ipmi.ChassisControlPowerDown
	}

	if _, err := client.ChassisControl(ctx, cont); err != nil {
		panic(err)
	}
	return true
}

func SendMagicPacket(macAddr string) error {
	hw, err := net.ParseMAC(macAddr)
	if err != nil {
		return err
	}

	packet := bytes.Repeat([]byte{0xFF}, 6)
	packet = append(packet, bytes.Repeat(hw, 16)...)

	broadcastAddr := net.UDPAddr{IP: net.IPv4bcast, Port: 9}
	conn, err := net.DialUDP("udp", nil, &broadcastAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}
