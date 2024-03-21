package main

import (
	"encoding/json"
	"os"
)

type ConfigRunner struct {
	Name                string  `json:"name"`
	IOType              string  `json:"io-type"`
	MqUrl               string  `json:"mq-url"`
	MqInputQueue        string  `json:"mq-input-queue"`
	MqOutputExchange    string  `json:"mq-output-exchange"`
	MqOutputRoutingKey  string  `json:"mq-output-routing-key"`
	FileInputPath       string  `json:"file-input-path"`
	FileOutputPath      string  `json:"file-output-path"`
	FileRestartCount    int     `json:"file-restart-count"`
	FileRestartInterval float32 `json:"file-restart-interval"`
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
