package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

type DeviceConfig struct {
	Name         string  `json:"name"`
	DeviceClass  *string `json:"device_class,omitempty"`
	StateTopic   *string `json:"state_topic,omitempty"`
	CommandTopic *string `json:"command_topic,omitempty"`
	UniqueID     string  `json:"unique_id"`
	Device       Device  `json:"device"`
}

type Device struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
}

type BasicServer struct {
	Name     string
	UniqueID string

	TopicStopConfig  string
	TopicStopCommand string

	TopicStartConfig  string
	TopicStartCommand string

	TopicPowerState  string
	TopicPowerConfig string

	Host string
}

func (s *BasicServer) Discovery(client mqtt.Client) {
	send := func(topic string, cfg DeviceConfig) {
		jsonPayload, err := json.Marshal(cfg)
		if err != nil {
			log.Fatalf("JSON marshaling failed: %v", err)
		}

		client.Publish(topic, 0, false, jsonPayload)

		log.Println(topic)
	}

	deviceInfo := Device{
		Identifiers:  []string{s.UniqueID},
		Name:         s.Name,
		Manufacturer: "SiRMonkeys",
		Model:        "BasicServer",
	}

	runningStr := "running"
	send(s.TopicPowerConfig, DeviceConfig{
		Name:        "State",
		DeviceClass: &runningStr,
		StateTopic:  &s.TopicPowerState,
		UniqueID:    s.UniqueID + "_power",
		Device:      deviceInfo,
	})

	send(s.TopicStopConfig, DeviceConfig{
		Name:         "Stop",
		CommandTopic: &s.TopicStopCommand,
		UniqueID:     s.UniqueID + "_stop",
		Device:       deviceInfo,
	})
	client.Subscribe(s.TopicStopCommand, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("MSG: Stop Server %s: %s\n", s.UniqueID, string(msg.Payload()))
		s.Stop(client)
	})

	send(s.TopicStartConfig, DeviceConfig{
		Name:         "Start",
		CommandTopic: &s.TopicStartCommand,
		UniqueID:     s.UniqueID + "_start",
		Device:       deviceInfo,
	})
	client.Subscribe(s.TopicStartCommand, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("MSG: Start Server %s: %s\n", s.UniqueID, string(msg.Payload()))
		s.Start(client)
	})
}

func (s *BasicServer) Stop(client mqtt.Client) {
	client.Publish(s.TopicPowerState, 0, false, "UNKNOWN")
	PowerIpmi("10.1.1.239", "ADMIN", "ADMIN", "OFF")
}

func (s *BasicServer) Start(client mqtt.Client) {
	client.Publish(s.TopicPowerState, 0, false, "UNKNOWN")
	PowerIpmi("10.1.1.239", "ADMIN", "ADMIN", "ON")
}

func (s *BasicServer) Check(client mqtt.Client) {
	log.Printf("Checking %s\n", s.Host)

	state := IsServerUpHTTP("http://"+s.Host, 30*time.Second)

	stateStr := "OFF"
	if state {
		stateStr = "ON"
	}

	log.Printf("State %s: %s\n", s.Host, stateStr)

	client.Publish(s.TopicPowerState, 0, false, stateStr)
}

func NewBasicServer(uuid string, name string, host string) *BasicServer {
	prefix := "homeassistant"
	s := &BasicServer{
		Name:              name,
		Host:              host,
		UniqueID:          uuid,
		TopicPowerState:   fmt.Sprintf("%s/binary_sensor/%s/%s/state", prefix, uuid, "power"),
		TopicPowerConfig:  fmt.Sprintf("%s/binary_sensor/%s/%s/config", prefix, uuid, "power"),
		TopicStopConfig:   fmt.Sprintf("%s/button/%s/%s/config", prefix, uuid, "stop"),
		TopicStopCommand:  fmt.Sprintf("%s/button/%s/%s/command", prefix, uuid, "stop"),
		TopicStartConfig:  fmt.Sprintf("%s/button/%s/%s/config", prefix, uuid, "start"),
		TopicStartCommand: fmt.Sprintf("%s/button/%s/%s/command", prefix, uuid, "start"),
	}

	return s
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, relying on environment variables")
	}

	opts := mqtt.NewClientOptions().
		AddBroker(os.Getenv("MQTT_BROKER")).
		SetClientID(os.Getenv("MQTT_CLIENT_ID")).
		SetUsername(os.Getenv("MQTT_USERNAME")).
		SetPassword(os.Getenv("MQTT_PASSWORD")).
		SetDefaultPublishHandler(messageHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT connection failed: %v", token.Error())
	}

	log.Printf("Connected to MQTT broker at %s\n\n", os.Getenv("MQTT_BROKER"))

	s := NewBasicServer("ccf337ed-a7b9-4b26-afa5-51fac9a56ccb", "TrueNas Server", "10.1.1.20")
	s.Discovery(client)
	s.Check(client)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.Check(client)
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
