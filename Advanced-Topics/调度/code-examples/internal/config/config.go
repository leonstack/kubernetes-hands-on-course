// Package config provides configuration management for scheduler tools
package config

import (
	"fmt"
	"os"
	"time"

	"sigs.k8s.io/yaml"
)

// CommonConfig holds common configuration for all scheduler tools
type CommonConfig struct {
	LogLevel    string        `yaml:"log_level" json:"log_level"`
	MetricsPort int           `yaml:"metrics_port" json:"metrics_port"`
	HealthPort  int           `yaml:"health_port" json:"health_port"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	Kubeconfig  string        `yaml:"kubeconfig" json:"kubeconfig"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*CommonConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config CommonConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Set defaults
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.MetricsPort == 0 {
		config.MetricsPort = 8080
	}
	if config.HealthPort == 0 {
		config.HealthPort = 8081
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &config, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *CommonConfig {
	return &CommonConfig{
		LogLevel:    "info",
		MetricsPort: 8080,
		HealthPort:  8081,
		Timeout:     30 * time.Second,
		Kubeconfig:  "", // Will use in-cluster config if empty
	}
}
