package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type EnvConfig struct {
	MQTT_BROKER    string `env:"MQTT_BROKER" required:"true"`
	MQTT_CLIENT_ID string `env:"MQTT_CLIENT_ID" required:"true"`
	MQTT_USERNAME  string `env:"MQTT_USERNAME" required:"true"`
	MQTT_PASSWORD  string `env:"MQTT_PASSWORD" required:"true"`

	CONFIG_FILE string `env:"CONFIG_FILE"`
}

func loadEnv() EnvConfig {
	godotenv.Load()
	var cfg EnvConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to process env vars: %v", err)
	}
	return cfg
}

func ResolveConfigPath(explicit string) string {
	if explicit != "" {
		return explicit
	}

	if _, err := os.Stat("servers.yaml"); err == nil {
		return "servers.yaml"
	}

	if _, err := os.Stat("server.yml"); err == nil {
		log.Println("CONFIG_FILE not set, using legacy config path server.yml")
		return "server.yml"
	}

	return "servers.yaml"
}

// Server.yml config
type Action struct {
	Type   string   `yaml:"type"`
	Params []string `yaml:"params"`
}

type Server struct {
	UUID  string `yaml:"uuid"`
	Name  string `yaml:"name"`
	Check Action `yaml:"check"`
	Start Action `yaml:"start"`
	Stop  Action `yaml:"stop"`
}

type Config struct {
	Servers []Server `yaml:"servers"`
}

func LoadServers(filename string) Config {
	var cfg Config
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = Config{
				Servers: []Server{
					{
						UUID:  "default-uuid",
						Name:  "default-server",
						Check: Action{Type: "ping", Params: []string{"127.0.0.1"}},
						Start: Action{Type: "wol", Params: []string{"00:00:00:00:00:00"}},
						Stop:  Action{Type: "ssh", Params: []string{"127.0.0.1", "shutdown now"}},
					},
				},
			}
			log.Println("No config file found, using default configuration.")
			if err := SaveServers(filename, cfg); err != nil {
				log.Printf("Error saving default config: %v\n", err)
			}
			return cfg
		}
		panic(err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	for _, server := range cfg.Servers {
		log.Printf("Server: %s (%s)\n", server.Name, server.UUID)
		log.Printf("  Check: %s %v\n", server.Check.Type, server.Check.Params)
		log.Printf("  Start: %s %v\n", server.Start.Type, server.Start.Params)
		log.Printf("  Stop : %s %v\n", server.Stop.Type, server.Stop.Params)
	}
	return cfg
}

func SaveServers(filename string, cfg Config) error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}

func ValidateServers(cfg Config) error {
	var errs []string

	for idx, server := range cfg.Servers {
		prefix := fmt.Sprintf("servers[%d]", idx)

		if strings.TrimSpace(server.UUID) == "" {
			errs = append(errs, fmt.Sprintf("%s: uuid is required", prefix))
		}
		if strings.TrimSpace(server.Name) == "" {
			errs = append(errs, fmt.Sprintf("%s: name is required", prefix))
		}

		errs = append(errs, validateAction(prefix, "check", server.Check)...)
		errs = append(errs, validateAction(prefix, "start", server.Start)...)
		errs = append(errs, validateAction(prefix, "stop", server.Stop)...)
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid server config:\n- %s", strings.Join(errs, "\n- "))
	}
	return nil
}

func validateAction(serverPrefix, actionName string, action Action) []string {
	var errs []string

	actionType := strings.TrimSpace(action.Type)
	if actionType == "" {
		return []string{fmt.Sprintf("%s.%s: type is required", serverPrefix, actionName)}
	}

	requiredParams, known := builtinActionParams(actionName, actionType)
	if known {
		if len(action.Params) != requiredParams {
			errs = append(errs, fmt.Sprintf("%s.%s: type %q requires %d params, got %d", serverPrefix, actionName, actionType, requiredParams, len(action.Params)))
		}
		return errs
	}

	if _, found := FindLuaScript(actionType); !found {
		errs = append(errs, fmt.Sprintf("%s.%s: unsupported type %q", serverPrefix, actionName, actionType))
	}

	return errs
}

func builtinActionParams(actionName, actionType string) (int, bool) {
	switch actionName {
	case "check":
		switch actionType {
		case "ping", "http":
			return 1, true
		}
	case "start":
		switch actionType {
		case "wol":
			return 1, true
		case "ipmi":
			return 3, true
		}
	case "stop":
		switch actionType {
		case "ssh":
			return 2, true
		case "ipmi":
			return 3, true
		}
	}

	return 0, false
}
