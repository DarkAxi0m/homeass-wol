package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	ProductName = "HomeAss-Wol"
)

func main() {
	cfg := loadEnv()

	serverscfg := LoadServers(cfg.CONFIG_FILE)

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT_BROKER).
		SetClientID(cfg.MQTT_CLIENT_ID).
		SetUsername(cfg.MQTT_USERNAME).
		SetPassword(cfg.MQTT_PASSWORD).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(30 * time.Second).
		SetDefaultPublishHandler(messageHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT connection failed: %v", token.Error())
	}

	log.Printf("Connected to MQTT broker at %s\n\n", cfg.MQTT_BROKER)

	resultChan := make(chan *BasicServer, len(serverscfg.Servers))

	for _, servercfg := range serverscfg.Servers {
		go func(cfg Server) {
			s := NewBasicServer(cfg)
			s.Discovery(client)
			s.Check(client)
			resultChan <- s
		}(servercfg)
	}

	var servers []*BasicServer
	for range serverscfg.Servers {
		servers = append(servers, <-resultChan)
	}

	// Main Checking Loop
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, s := range servers {
			s.Check(client)
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	client.Disconnect(250)
	log.Println("Disconnected from MQTT broker")
}

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message on topic %s: %s\n", msg.Topic(), string(msg.Payload()))
}
