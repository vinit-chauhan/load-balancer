package config

import (
	"fmt"
	"os"

	"github.com/vinit-chauhan/reverse-proxy/logger"
	"gopkg.in/yaml.v3"
)

var (
	ConfigPath = "./config.yml"
	config     = ConfigType{}
)

type ConfigType struct {
	Services []ServiceType `yaml:"services"`
}

type ServiceType struct {
	Name     string   `yaml:"name"`
	Backends []string `yaml:"urls"`
	UrlPath  string   `yaml:"url_path"`
}

func Load() {
	buff, err := os.ReadFile(ConfigPath)
	if err != nil {
		logger.Panic("Load", fmt.Sprintf("Error loading config file from disk: %s: %s", ConfigPath, err.Error()))
	}

	if err := yaml.Unmarshal(buff, &config); err != nil {
		logger.Panic("Load", "error unmarshaling config file:"+err.Error())
	}
}

func (s *ServiceType) Validate() {
	if s.UrlPath[0] != '/' {
		logger.Error("Validate", "error URLPath must start with '/'")
		panic("validation error: URLPath: " + s.UrlPath)
	}
}

func GetConfig() ConfigType {
	return config
}
