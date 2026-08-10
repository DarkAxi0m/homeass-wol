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

var mqttClient mqtt.Client

func main() {
	cfg := loadEnv()
	cfg.CONFIG_FILE = ResolveConfigPath(cfg.CONFIG_FILE)

	serverscfg := LoadServers(cfg.CONFIG_FILE)
	if err := ValidateServers(serverscfg); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT_BROKER).
		SetClientID(cfg.MQTT_CLIENT_ID).
		SetUsername(cfg.MQTT_USERNAME).
		SetPassword(cfg.MQTT_PASSWORD).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(30 * time.Second)
		//		SetDefaultPublishHandler(messageHandler)

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT connection failed: %v", token.Error())
	}

	log.Printf("Connected to MQTT broker at %s\n\n", cfg.MQTT_BROKER)

	resultChan := make(chan *BasicServer, len(serverscfg.Servers))

	for _, servercfg := range serverscfg.Servers {
		go func(cfg Server) {
			s := NewBasicServer(cfg)
			s.Discovery()
			s.Check()
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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	runMainLoop(ticker.C, sig, servers)

	mqttClient.Disconnect(250)
	log.Println("Disconnected from MQTT broker")
}

func runMainLoop(ticks <-chan time.Time, sig <-chan os.Signal, servers []*BasicServer) {
	for {
		select {
		case <-ticks:
			for _, s := range servers {
				s.Check()
			}
		case <-sig:
			return
		}
	}
}
