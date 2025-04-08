package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"mqtt-go-app/homeassistant"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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

	lastState string
	cfg       Server
}

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

// Need to learn more..., this seem "wrong"
func StringPtr(s string) *string {
	return &s
}

func (s *BasicServer) SetState(to bool) {
	stateStr := "OFF"
	if to {
		stateStr = "ON"
	}

	if stateStr != s.lastState {
		s.lastState = stateStr
	}

	log.Printf("Check for %s: %s \n", s.Name, stateStr)
	mqttClient.Publish(s.TopicPowerState, 0, false, stateStr)
	if to {
		mqttClient.Publish(s.TopicLastSeenState, 0, false, time.Now().Format(time.RFC3339))
	}
}

func (s *BasicServer) Discovery() {
	send := func(topic string, cfg DeviceConfig) {
		jsonPayload, err := json.Marshal(cfg)
		if err != nil {
			log.Fatalf("JSON marshaling failed: %v", err)
		}

		mqttClient.Publish(topic, 0, true, jsonPayload)

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
	mqttClient.Subscribe(s.TopicStopCommand, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("MSG: Stop Server %s: %s\n", s.UniqueID, string(msg.Payload()))
		mqttClient.Publish(s.TopicPowerState, 0, false, "UNKNOWN")
		s.Stop()
	})

	send(s.TopicStartConfig, DeviceConfig{
		Name:         "Start",
		CommandTopic: &s.TopicStartCommand,
		UniqueID:     s.UniqueID + "_start",
		Device:       deviceInfo,
	})
	mqttClient.Subscribe(s.TopicStartCommand, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("MSG: Start Server %s: %s\n", s.UniqueID, string(msg.Payload()))
		mqttClient.Publish(s.TopicPowerState, 0, false, "UNKNOWN")
		s.Start()
	})
}

func luaCheck(name string, action Action) (bool, bool) {
	if result, found := RunLuaScript(action.Type, action.Params); found {
		log.Printf("[LUA] %s %s (%s): %s", action.Type, name, action.Params, result)
		return strings.TrimSpace(result) == "success", true
	}
	return false, false
}

func (s *BasicServer) Check() {
	if result, found := luaCheck(s.Name, s.cfg.Check); found {
		s.SetState(result)
	} else {
		fmt.Printf("Unknown Check: %s, %s.\n", s.Name, s.cfg.Check.Type)
	}
}

func (s *BasicServer) Start() bool {
	if result, found := luaCheck(s.Name, s.cfg.Start); found {
		return result
	}

	switch t := s.cfg.Start.Type; t {
	case "wol":
		mac := s.cfg.Start.Params[0]
		if err := SendMagicPacket(mac); err != nil {
			fmt.Println("Error:", err, mac)
		} else {
			fmt.Println("Magic packet sent!", mac)
		}
	case "ipmi":
		params := s.cfg.Start.Params
		return PowerIpmi(params[0], params[1], params[2], "ON")

	default:
		fmt.Printf("Unknown start: %s.\n", t)
	}
	return false
}

func (s *BasicServer) Stop() bool {
	if result, found := luaCheck(s.Name, s.cfg.Stop); found {
		return result
	}

	switch t := s.cfg.Stop.Type; t {
	case "ssh":
		params := s.cfg.Stop.Params
		error := RunSSHCommand(params[0], params[1])
		if error != nil {
			fmt.Println("SSH Run error", error)
			return false
		}
		return true
	case "ipmi":
		params := s.cfg.Stop.Params
		return PowerIpmi(params[0], params[1], params[2], "OFF")

	default:
		fmt.Printf("Unknown Stop: %s.\n", t)
	}
	return false
}

func NewBasicServer(servercfg Server) *BasicServer {
	uuid := servercfg.UUID

	prefix := "homeassistant"
	s := &BasicServer{
		Name:                servercfg.Name,
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

		cfg: servercfg,

		lastState: "Off",
	}

	return s
}
