package main

import (
	"encoding/json"
	"os"
)

type ConfigRunner struct {
	Name string `json:"name"`
}

type ConfigRunners struct {
	Services []ConfigRunner `json:"services"`
}

func (c *ConfigRunners) LoadConfigFromFile(configPath string) error {
	byteResult, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(byteResult), c)
}

func (c *ConfigRunners) GetConfig(service *LocalService) (*ConfigRunner, bool) {
	for _, s := range c.Services {
		if service.Name == s.Name {
			return &s, true
		}
	}
	return nil, false
}
