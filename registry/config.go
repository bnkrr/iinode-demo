package main

// @Title        config.go
// @Description  用于读取配置文件并管理配置
// @Create       dlchang (2024/04/03 15:00)

import (
	"encoding/json"
	"os"
	"strings"
)

type ConfigRunnerMatchResultType int

const (
	ConfigRunner_NOTFOUND ConfigRunnerMatchResultType = iota
	ConfigRunner_SAME
	ConfigRunner_DIFFERENT
)

type ConfigEtcd struct {
	Url      string `json:"url"`      // etcd服务端地址，形如 localhost:12345
	User     string `json:"user"`     // 用户名
	Password string `json:"password"` // 密码
	Prefix   string `json:"prefix"`   // key前缀
}

type ConfigMQ struct {
	Url                      string `json:"url"`                         // MQ的配置地址，形如 amqp://user:password@localhost:5672/vhost
	OutputExchange           string `json:"output-exchange"`             // MQ输出exchange名称
	InputQueuePrefix         string `json:"input-queue-prefix"`          // 输入队列前缀
	OutputRoutingKeyPrefix   string `json:"output-routing-key-prefix"`   // 输出key前缀
	AllowNoName              bool   `json:"allow-no-name"`               // 是否允许未被列出的service
	DefaultOutputQueuePrefix string `json:"default-output-queue-prefix"` // 未被列出service默认输出队列前缀
	Version                  string `json:"version"`                     // 配置的版本，用于区分自动更新配置
}

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
	Mq1NextService      string  `json:"mq1-next-service"`      // 下一个服务名
	Mq1OutputQueue      string  `json:"mq1-output-queue"`      // 或直接输出到队列
	Version             string  `json:"version"`               // 配置的版本，用于区分自动更新配置
}

type ConfigRegistry struct {
	Etcd     ConfigEtcd     `json:"etcd"`
	RabbitMQ ConfigMQ       `json:"rabbitmq"`
	Services []ConfigRunner `json:"services"`
}

func (c *ConfigRegistry) InitIOType() error {
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

func (c *ConfigRegistry) MQEnabled() bool {
	return c.RabbitMQ.Url != ""
}

func (c *ConfigRegistry) LoadConfigFromFile(configPath string) error {
	byteResult, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	return c.LoadConfigFromBytes([]byte(byteResult))
}

func (c *ConfigRegistry) LoadConfigFromBytes(cfgbytes []byte) error {
	err := json.Unmarshal(cfgbytes, c)
	if err != nil {
		return err
	}
	return c.InitIOType()
}

// @title        GetConfig
// @description  查询本地服务service的对应配置
// @auth         dlchang (2024/04/03 15:17)
// @param        serviceName   string   服务名称
// @return       *ConfigRunner   对应服务的runner配置，查询不到则为nil
// @return       bool   是否可以查询到对应的runner配置
func (c *ConfigRegistry) GetConfigRunner(serviceName string) (*ConfigRunner, bool) {
	for _, s := range c.Services {
		if serviceName == s.Name {
			return &s, true
		}
	}
	if c.RabbitMQ.AllowNoName {
		return &ConfigRunner{
			Name:           serviceName,
			Mq1OutputQueue: c.RabbitMQ.DefaultOutputQueuePrefix + strings.ToLower(serviceName),
			Version:        c.RabbitMQ.Version,
		}, true
	}
	return nil, false
}

func (c *ConfigRegistry) MatchConfigRunner(configRunner *ConfigRunner) ConfigRunnerMatchResultType {
	cfg, ok := c.GetConfigRunner(configRunner.Name)
	if !ok {
		return ConfigRunner_NOTFOUND
	}
	if cfg.Version == configRunner.Version {
		return ConfigRunner_SAME
	}
	return ConfigRunner_DIFFERENT
}
