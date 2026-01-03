package app

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func InitDB(path string) error {
	// Verify database file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}
