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

var (
	runLuaScriptFunc    = RunLuaScript
	sendMagicPacketFunc = SendMagicPacket
	runSSHCommandFunc   = RunSSHCommand
	powerIPMIFunc       = PowerIpmi
	publishMQTTFunc     = func(topic string, payload any, retained bool) {
		if mqttClient == nil {
			return
		}
		mqttClient.Publish(topic, 0, retained, payload)
	}
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
	publishMQTT(s.TopicPowerState, stateStr, false)
	if to {
		publishMQTT(s.TopicLastSeenState, time.Now().Format(time.RFC3339), false)
	}
}

func (s *BasicServer) PublishUnknownState() {
	publishMQTT(s.TopicPowerState, "UNKNOWN", false)
}

func publishMQTT(topic string, payload any, retained bool) {
	publishMQTTFunc(topic, payload, retained)
}

func (s *BasicServer) Discovery() {
	send := func(topic string, cfg DeviceConfig) {
		jsonPayload, err := json.Marshal(cfg)
		if err != nil {
			log.Fatalf("JSON marshaling failed: %v", err)
		}

		publishMQTT(topic, jsonPayload, true)

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
	if mqttClient != nil {
		mqttClient.Subscribe(s.TopicStopCommand, 0, func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("MSG: Stop Server %s: %s\n", s.UniqueID, string(msg.Payload()))
			s.PublishUnknownState()
			s.Stop()
		})
	}

	send(s.TopicStartConfig, DeviceConfig{
		Name:         "Start",
		CommandTopic: &s.TopicStartCommand,
		UniqueID:     s.UniqueID + "_start",
		Device:       deviceInfo,
	})
	if mqttClient != nil {
		mqttClient.Subscribe(s.TopicStartCommand, 0, func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("MSG: Start Server %s: %s\n", s.UniqueID, string(msg.Payload()))
			s.PublishUnknownState()
			s.Start()
		})
	}
}

func luaCheck(name string, action Action) (bool, bool, error) {
	if result, found, err := runLuaScriptFunc(action.Type, action.Params); found {
		if err != nil {
			return false, true, fmt.Errorf("[LUA] %s %s: %w", action.Type, name, err)
		}
		log.Printf("[LUA] %s %s (%v): %s", action.Type, name, action.Params, result)
		return strings.TrimSpace(result) == "success", true, nil
	}
	return false, false, nil
}

func (s *BasicServer) Check() {
	if result, found, err := luaCheck(s.Name, s.cfg.Check); err != nil {
		log.Printf("Check failed for %s: %v", s.Name, err)
		s.PublishUnknownState()
	} else if found {
		s.SetState(result)
	} else {
		log.Printf("Unsupported check type for %s: %s", s.Name, s.cfg.Check.Type)
	}
}

func (s *BasicServer) Start() bool {
	if result, found, err := luaCheck(s.Name, s.cfg.Start); err != nil {
		log.Printf("Start failed for %s: %v", s.Name, err)
		return false
	} else if found {
		return result
	}

	switch t := s.cfg.Start.Type; t {
	case "wol":
		if len(s.cfg.Start.Params) < 1 {
			log.Printf("Start failed for %s: wol requires 1 param", s.Name)
			return false
		}
		mac := s.cfg.Start.Params[0]
		if err := sendMagicPacketFunc(mac); err != nil {
			log.Printf("WOL failed for %s: %v", s.Name, err)
			return false
		}
		log.Printf("Magic packet sent for %s: %s", s.Name, mac)
		return true
	case "ipmi":
		params := s.cfg.Start.Params
		if len(params) < 3 {
			log.Printf("Start failed for %s: ipmi requires 3 params", s.Name)
			return false
		}
		if err := powerIPMIFunc(params[0], params[1], params[2], "ON"); err != nil {
			log.Printf("IPMI start failed for %s: %v", s.Name, err)
			return false
		}
		return true

	default:
		log.Printf("Unsupported start type for %s: %s", s.Name, t)
	}
	return false
}

func (s *BasicServer) Stop() bool {
	if result, found, err := luaCheck(s.Name, s.cfg.Stop); err != nil {
		log.Printf("Stop failed for %s: %v", s.Name, err)
		return false
	} else if found {
		return result
	}

	switch t := s.cfg.Stop.Type; t {
	case "ssh":
		params := s.cfg.Stop.Params
		if len(params) < 2 {
			log.Printf("Stop failed for %s: ssh requires 2 params", s.Name)
			return false
		}
		error := runSSHCommandFunc(params[0], params[1])
		if error != nil {
			log.Printf("SSH stop failed for %s: %v", s.Name, error)
			return false
		}
		return true
	case "ipmi":
		params := s.cfg.Stop.Params
		if len(params) < 3 {
			log.Printf("Stop failed for %s: ipmi requires 3 params", s.Name)
			return false
		}
		if err := powerIPMIFunc(params[0], params[1], params[2], "OFF"); err != nil {
			log.Printf("IPMI stop failed for %s: %v", s.Name, err)
			return false
		}
		return true

	default:
		log.Printf("Unsupported stop type for %s: %s", s.Name, t)
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
