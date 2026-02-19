package config

import (
	"os"

	"github.com/vinit-chauhan/load-balancer/logger"
	"gopkg.in/yaml.v3"
)

var (
	config = ConfigType{}
)

type ConfigType struct {
	Services []ServiceType `yaml:"services"`
}

type ServiceType struct {
	Name        string            `yaml:"name"`
	Backends    []string          `yaml:"urls"`
	UrlPath     string            `yaml:"endpoint"`
	Algorithm   string            `yaml:"algorithm"`    // "round-robin", "least-connections", "ip-hash"
	HealthCheck HealthCheckConfig `yaml:"health_check"`
}

type HealthCheckConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
	Path     string `yaml:"path"`
}

func Load(path string) {
	buff, err := os.ReadFile(path)
	if err != nil {
		logger.Panic("Load", "Error loading config file from disk", "path", path, "error", err)
	}

	if err := yaml.Unmarshal(buff, &config); err != nil {
		logger.Panic("Load", "Error unmarshaling config file", "error", err)
	}
}

func (s *ServiceType) Validate() {
	if s.UrlPath[0] != '/' {
		logger.Error("Validate", "error URLPath must start with '/'")
		panic("validation error: URLPath: " + s.UrlPath)
	}
	if s.Algorithm == "" {
		s.Algorithm = "round-robin"
	}
}

func GetConfig() ConfigType {
	return config
}
