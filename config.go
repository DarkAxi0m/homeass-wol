package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

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

func LoadConfig(filename string) Config {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		panic(err)
	}

	for _, server := range cfg.Servers {
		fmt.Printf("Server: %s (%s)\n", server.Name, server.UUID)
		fmt.Printf("  Check: %s %v\n", server.Check.Type, server.Check.Params)
		fmt.Printf("  Start: %s %v\n", server.Start.Type, server.Start.Params)
		fmt.Printf("  Stop : %s %v\n", server.Stop.Type, server.Stop.Params)
	}
	return cfg
}
