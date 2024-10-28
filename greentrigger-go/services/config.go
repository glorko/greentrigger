package services

import (
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"os"
)

type ServiceConfig struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Command string            `yaml:"command"`
	Env     map[string]string `yaml:"env"`
}

type Config struct {
	Services []ServiceConfig `yaml:"services"`
}

// LoadConfig reads the YAML config file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &config, nil
}
