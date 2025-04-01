package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/bougou/go-ipmi"
)

func RunSSHCommand(host, command string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "ssh", host, command).CombinedOutput()
	if err != nil {
		return fmt.Errorf("SSH command error: %s, output: %s", err.Error(), string(out))
	}
	return nil
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
