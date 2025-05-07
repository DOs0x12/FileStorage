package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	KafkaAddress   string `yaml:"kafkaAddress"`
	StorageAddress string `yaml:"storageAddress"`
	StorageDB      string `yaml:"storageDB"`
	StorageUser    string `yaml:"storageUser"`
	StoragePass    string `yaml:"storagePass"`
}

func LoadConfig(path string) (Config, error) {
	confFile, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read the config file: %w", err)
	}

	return unmarshalConf(confFile)
}

func unmarshalConf(data []byte) (Config, error) {
	dto := Config{}
	err := yaml.Unmarshal(data, &dto)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal the config data: %w", err)
	}

	return dto, nil
}
