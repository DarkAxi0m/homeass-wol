package main

import (
	"errors"
	"testing"
)

func TestBasicServerStart(t *testing.T) {
	origSendMagicPacket := sendMagicPacketFunc
	origPowerIPMI := powerIPMIFunc
	origRunLua := runLuaScriptFunc
	t.Cleanup(func() {
		sendMagicPacketFunc = origSendMagicPacket
		powerIPMIFunc = origPowerIPMI
		runLuaScriptFunc = origRunLua
	})

	t.Run("wol success returns true", func(t *testing.T) {
		runLuaScriptFunc = func(name string, params []string) (string, bool, error) {
			return "", false, nil
		}
		sendMagicPacketFunc = func(mac string) error { return nil }

		s := &BasicServer{
			Name: "server",
			cfg: Server{
				Start: Action{Type: "wol", Params: []string{"00:11:22:33:44:55"}},
			},
		}

		if got := s.Start(); !got {
			t.Fatal("Start() = false, want true")
		}
	})

	t.Run("wol failure returns false", func(t *testing.T) {
		runLuaScriptFunc = func(name string, params []string) (string, bool, error) {
			return "", false, nil
		}
		sendMagicPacketFunc = func(mac string) error { return errors.New("send failed") }

		s := &BasicServer{
			Name: "server",
			cfg: Server{
				Start: Action{Type: "wol", Params: []string{"00:11:22:33:44:55"}},
			},
		}

		if got := s.Start(); got {
			t.Fatal("Start() = true, want false")
		}
	})

	t.Run("ipmi failure returns false", func(t *testing.T) {
		runLuaScriptFunc = func(name string, params []string) (string, bool, error) {
			return "", false, nil
		}
		powerIPMIFunc = func(host, username, password, state string) error {
			return errors.New("ipmi failed")
		}

		s := &BasicServer{
			Name: "server",
			cfg: Server{
				Start: Action{Type: "ipmi", Params: []string{"host", "user", "pass"}},
			},
		}

		if got := s.Start(); got {
			t.Fatal("Start() = true, want false")
		}
	})

	t.Run("lua execution error returns false", func(t *testing.T) {
		runLuaScriptFunc = func(name string, params []string) (string, bool, error) {
			return "", true, errors.New("lua failure")
		}

		s := &BasicServer{
			Name: "server",
			cfg: Server{
				Start: Action{Type: "custom"},
			},
		}

		if got := s.Start(); got {
			t.Fatal("Start() = true, want false")
		}
	})
}

func TestBasicServerCheck(t *testing.T) {
	origRunLua := runLuaScriptFunc
	origPublish := publishMQTTFunc
	t.Cleanup(func() {
		runLuaScriptFunc = origRunLua
		publishMQTTFunc = origPublish
	})

	t.Run("lua check error publishes unknown", func(t *testing.T) {
		var published []string
		runLuaScriptFunc = func(name string, params []string) (string, bool, error) {
			return "", true, errors.New("lua failure")
		}
		publishMQTTFunc = func(topic string, payload any, retained bool) {
			if value, ok := payload.(string); ok {
				published = append(published, value)
			}
		}

		s := &BasicServer{
			Name:            "server",
			TopicPowerState: "power/topic",
			lastState:       "ON",
			cfg: Server{
				Check: Action{Type: "custom"},
			},
		}

		s.Check()

		if len(published) != 1 || published[0] != "UNKNOWN" {
			t.Fatalf("published = %v, want [UNKNOWN]", published)
		}
		if s.lastState != "ON" {
			t.Fatalf("lastState = %q, want unchanged ON", s.lastState)
		}
	})

	t.Run("unsupported check does not mutate state", func(t *testing.T) {
		var publishCount int
		runLuaScriptFunc = func(name string, params []string) (string, bool, error) {
			return "", false, nil
		}
		publishMQTTFunc = func(topic string, payload any, retained bool) {
			publishCount++
		}

		s := &BasicServer{
			Name:      "server",
			lastState: "ON",
			cfg: Server{
				Check: Action{Type: "missing"},
			},
		}

		s.Check()

		if publishCount != 0 {
			t.Fatalf("publishCount = %d, want 0", publishCount)
		}
		if s.lastState != "ON" {
			t.Fatalf("lastState = %q, want unchanged ON", s.lastState)
		}
	})
}
