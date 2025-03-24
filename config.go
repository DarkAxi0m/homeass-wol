package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type EnvConfig struct {
	MQTT_BROKER    string `env:"MQTT_BROKER" required:"true"`
	MQTT_CLIENT_ID string `env:"MQTT_CLIENT_ID" required:"true"`
	MQTT_USERNAME  string `env:"MQTT_USERNAME" required:"true"`
	MQTT_PASSWORD  string `env:"MQTT_PASSWORD" required:"true"`

	CONFIG_FILE string `env:"CONFIG_FILE" default:"server.yml"`
}

func loadEnv() EnvConfig {
	godotenv.Load()
	var cfg EnvConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to process env vars: %v", err)
	}
	return cfg
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
						Start: Action{Type: "systemctl", Params: []string{"start", "service"}},
						Stop:  Action{Type: "systemctl", Params: []string{"stop", "service"}},
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
