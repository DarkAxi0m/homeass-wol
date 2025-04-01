package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mqtt-go-app/homeassistant"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	ProductName = "HomeAss-Wol"
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
	Model    string
	UniqueID string

	TopicStopConfig  string
	TopicStopCommand string

	TopicStartConfig  string
	TopicStartCommand string

	TopicPowerState  string
	TopicPowerConfig string

	TopicLastSeenState  string
	TopicLastSeenConfig string

	HealthCheck ServerHealthCheck
	Start       ServerStart
	Stop        ServerStop
}

type (
	ServerHealthCheck func(server *BasicServer) bool
	ServerStop        func(server *BasicServer) bool
	ServerStart       func(server *BasicServer) bool
)

// Need to learn more..., this seem "wrong"
func StringPtr(s string) *string {
	return &s
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
		Manufacturer: ProductName,
		Model:        s.Model,
	}

	send(s.TopicPowerConfig, DeviceConfig{
		Name:        "State",
		DeviceClass: StringPtr(string(homeassistant.BinarySensorClassRunning)),
		StateTopic:  &s.TopicPowerState,
		UniqueID:    s.UniqueID + "_power",
		Device:      deviceInfo,
	})
	send(s.TopicLastSeenConfig, DeviceConfig{
		Name:        "Last Seen",
		DeviceClass: StringPtr(string(homeassistant.SensorClassTimestamp)),
		StateTopic:  &s.TopicLastSeenState,
		UniqueID:    s.UniqueID + "_lastseen",
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
		client.Publish(s.TopicPowerState, 0, false, "UNKNOWN")
		if s.Stop != nil {
			s.Stop(s)
		}
	})

	send(s.TopicStartConfig, DeviceConfig{
		Name:         "Start",
		CommandTopic: &s.TopicStartCommand,
		UniqueID:     s.UniqueID + "_start",
		Device:       deviceInfo,
	})
	client.Subscribe(s.TopicStartCommand, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("MSG: Start Server %s: %s\n", s.UniqueID, string(msg.Payload()))
		client.Publish(s.TopicPowerState, 0, false, "UNKNOWN")
		if s.Start != nil {
			s.Start(s)
		}
	})
}

func (s *BasicServer) Check(client mqtt.Client) {
	state := false
	if s.HealthCheck != nil {
		state = s.HealthCheck(s)
	}
	stateStr := "OFF"
	if state {
		stateStr = "ON"
	}
	client.Publish(s.TopicPowerState, 0, false, stateStr)
	if state {
		client.Publish(s.TopicLastSeenState, 0, false, time.Now().Format(time.RFC3339))
	}
}

func NewBasicServer(uuid string, name string) *BasicServer {
	prefix := "homeassistant"
	s := &BasicServer{
		Name:                name,
		UniqueID:            uuid,
		Model:               "BasicServer",
		TopicPowerState:     fmt.Sprintf("%s/binary_sensor/%s/%s/state", prefix, uuid, "power"),
		TopicPowerConfig:    fmt.Sprintf("%s/binary_sensor/%s/%s/config", prefix, uuid, "power"),
		TopicLastSeenState:  fmt.Sprintf("%s/sensor/%s/%s/state", prefix, uuid, "lastseen"),
		TopicLastSeenConfig: fmt.Sprintf("%s/sensor/%s/%s/config", prefix, uuid, "lastseen"),
		TopicStopConfig:     fmt.Sprintf("%s/button/%s/%s/config", prefix, uuid, "stop"),
		TopicStopCommand:    fmt.Sprintf("%s/button/%s/%s/command", prefix, uuid, "stop"),
		TopicStartConfig:    fmt.Sprintf("%s/button/%s/%s/config", prefix, uuid, "start"),
		TopicStartCommand:   fmt.Sprintf("%s/button/%s/%s/command", prefix, uuid, "start"),
	}

	return s
}

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

	var servers []*BasicServer
	for _, servercfg := range serverscfg.Servers {
		s := NewBasicServer(servercfg.UUID, servercfg.Name)
		s.HealthCheck = ServerHealthCheck(func(server *BasicServer) bool {
			switch t := servercfg.Check.Type; t {
			case "pingcli":

				res, error := IsServerUpPingCli(servercfg.Check.Params[0], 29*time.Second)
				if error != nil {
					log.Print("Ping Error", error)
					return false
				}
				return res
			case "ping":

				return IsServerUpPing(servercfg.Check.Params[0], 29*time.Second)
			case "http":
				return IsServerUpHTTP(servercfg.Check.Params[0], 29*time.Second)

			default:
				fmt.Printf("Unknown Check: %s.\n", t)
			}
			return false
		})

		s.Start = ServerStart(func(server *BasicServer) bool {
			switch t := servercfg.Start.Type; t {
			case "wol":

				mac := servercfg.Start.Params[0]
				if err := SendMagicPacket(mac); err != nil {
					fmt.Println("Error:", err, mac)
				} else {
					fmt.Println("Magic packet sent!", mac)
				}
			case "ipmi":
				params := servercfg.Start.Params
				return PowerIpmi(params[0], params[1], params[2], "ON")

			default:
				fmt.Printf("Unknown start: %s.\n", t)
			}
			return false
		})

		s.Stop = ServerStop(func(server *BasicServer) bool {
			switch t := servercfg.Stop.Type; t {
			case "ssh":
				params := servercfg.Stop.Params
				error := RunSSHCommand(params[0], params[1])
				if error != nil {
					fmt.Println("SSH Run error", error)
					return false
				}
				return true
			case "ipmi":
				params := servercfg.Stop.Params
				return PowerIpmi(params[0], params[1], params[2], "OFF")

			default:
				fmt.Printf("Unknown Stop: %s.\n", t)
			}
			return false
		})

		s.Discovery(client)
		servers = append(servers, s)
	}
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
