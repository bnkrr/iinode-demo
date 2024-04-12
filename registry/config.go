package main

// @Title        config.go
// @Description  用于读取配置文件并管理配置
// @Create       dlchang (2024/04/03 15:00)

import (
	"encoding/json"
	"os"
)

type ConfigRunner struct {
	Name                string  `json:"name"`                  // 服务名称
	IOType              string  `json:"io-type"`               // IO类型
	InputType           string  `json:"input-type"`            // Input类型
	OutputType          string  `json:"output-type"`           // Output类型
	MqUrl               string  `json:"mq-url"`                // MQ的配置地址，形如 amqp://user:password@localhost:5672/vhost
	MqInputQueue        string  `json:"mq-input-queue"`        // MQ输入队列名称
	MqOutputExchange    string  `json:"mq-output-exchange"`    // MQ输出exchange名称
	MqOutputRoutingKey  string  `json:"mq-output-routing-key"` // MQ输出routing key名称（或输出队列名称）
	FileInputPath       string  `json:"file-input-path"`       // 文件方式输入文件路径
	FileOutputPath      string  `json:"file-output-path"`      // 文件方式输出文件路径（为空则不输出到文件）
	FileRestartCount    int     `json:"file-restart-count"`    // 文件方式重新发送输入的次数
	FileRestartInterval float32 `json:"file-restart-interval"` // 文件方式重新发送输入的时间间隔
}

type ConfigRunners struct {
	Services []ConfigRunner `json:"services"`
}

func (c *ConfigRunners) InitIOType() error {
	for _, cfg := range c.Services {
		if cfg.InputType == "" {
			cfg.InputType = cfg.IOType
		}
		if cfg.OutputType == "" {
			cfg.OutputType = cfg.IOType
		}
	}
	return nil
}

func (c *ConfigRunners) LoadConfigFromFile(configPath string) error {
	byteResult, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(byteResult), c)
	if err != nil {
		return err
	}
	return c.InitIOType()
}

// @title        GetConfig
// @description  查询本地服务service的对应配置
// @auth         dlchang (2024/04/03 15:17)
// @param        service   *LocalService   本地服务的指针，内涵服务名称等信息
// @return       *ConfigRunner   对应服务的runner配置，查询不到则为nil
// @return       bool   是否可以查询
func (c *ConfigRunners) GetConfig(service *LocalService) (*ConfigRunner, bool) {
	for _, s := range c.Services {
		if service.Name == s.Name {
			return &s, true
		}
	}
	return nil, false
}
