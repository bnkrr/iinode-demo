package main

import (
	"errors"
	"sync"
)

// @Title        iomanager.go
// @Description  实现统一的输入输出管理
// @Create       dlchang (2024/05/06 18:00)

// 目前有多种IO方式，runner不容易管理
// 引入iomanager统一管理多种输入输出方式，并实现针对service名称得到合适的输入输出

type IOManager struct {
	mqManager *MQConnectorManager
}

func (m *IOManager) GetInputConnector(c *ConfigRunner, wgroup *sync.WaitGroup) (InputConnector, error) {
	if c.InputType == "mq" {
		return NewIOConnectorMQ(c, wgroup)
	}
	if c.InputType == "file" {
		return NewIOConnectorFile(c, wgroup)
	}
	if m.mqManager != nil {
		return m.mqManager.GetNewInputConnector(c, wgroup)
	}
	return nil, errors.New("no io method matched")
}

func (m *IOManager) GetOutputConnector(c *ConfigRunner, wgroup *sync.WaitGroup) (OutputConnector, error) {
	if c.OutputType == "mq" {
		return NewIOConnectorMQ(c, wgroup)
	}
	if c.OutputType == "file" {
		return NewIOConnectorFile(c, wgroup)
	}
	if m.mqManager != nil {
		return m.mqManager.GetNewOutputConnector(c, wgroup)
	}
	return nil, errors.New("no io method matched")
}

func NewIOManager(cs *ConfigRunners) (*IOManager, error) {
	m := &IOManager{}
	if cs.MQEnabled() {
		mqManager, err := NewMQConnectorManger(cs.RabbitMQ)
		if err != nil {
			return nil, err
		}
		m.mqManager = mqManager
	} else {
		m.mqManager = nil
	}
	return m, nil
}
