package main

// @Title        ioconnectormq.go
// @Description  实现基于MQ队列的输入输出方式（用于正式测量）
// @Create       dlchang (2024/04/03 15:00)

// todo:
// 1. prefetch
// 2. 拆分IOConnectorMQ

// 目前MQ的connection还是按不同的runner分离设计，而不是统一的。
// 这样更灵活，方便支持小规模测绘，允许不同的service连接不同的MQ服务。
// 如果确定使用固定MQ接口，应该使用统一MQ服务，使用单个client，
// 利用两个connection分隔输入输出，建立多个channel和多个队列复用connection
import (
	"context"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MQClient struct {
	url     string
	conn    *amqp.Connection
	channel *amqp.Channel
}

func (m *MQClient) CheckInputQueue(inputQueue string) error {
	_, err := m.channel.QueueDeclarePassive(
		inputQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}

func (m *MQClient) GetInputQueue(inputQueue string) (<-chan amqp.Delivery, error) {
	return m.channel.Consume(
		inputQueue, // queue
		"",         // consumer
		false,      // autoAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,
	)
}

func (m *MQClient) CheckOutputQueue(outputExchange string, outputRoutingKey string) error {
	err := (error)(nil)
	if outputExchange != "" {
		err = m.channel.ExchangeDeclarePassive(
			outputExchange, // name
			"direct",       // type
			true,           // durable
			false,          // auto-deleted
			false,          // internal
			false,          // no-wait
			nil,            // arguments
		)
	} else {
		_, err = m.channel.QueueDeclarePassive(
			outputRoutingKey, // name
			true,             // durable
			false,            // delete when unused
			false,            // exclusive
			false,            // no-wait
			nil,              // arguments
		)
	}
	return err
}

func (m *MQClient) Publish(outputExchange string, outputRoutingKey string, output string) error {
	err := m.channel.PublishWithContext(context.Background(),
		outputExchange,   // exchange
		outputRoutingKey, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(output),
		})
	if err != nil {
		return err
	}
	return nil
}

func (m *MQClient) Close() {
	m.channel.Close()
	m.conn.Close()
}

func NewMQClient(url string) (*MQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &MQClient{
		url:     url,
		conn:    conn,
		channel: ch,
	}, nil
}

type InputMessageMQ struct {
	message amqp.Delivery
}

func (m *InputMessageMQ) Msg() string {
	return string(m.message.Body)
}

func (m *InputMessageMQ) Ack() {
	m.message.Ack(false)
}

type IOConnectorMQ struct {
	mqInputOnce  sync.Once
	mqOutputOnce sync.Once
	wgroup       *sync.WaitGroup
	config       *ConfigRunner
	mqClient     *MQClient
}

func (i *IOConnectorMQ) InputInit(ctx context.Context, inputCh chan InputMessage) {
	err := i.mqClient.CheckInputQueue(i.config.MqInputQueue)
	if err != nil {
		log.Panicf("%v", err)
	}
	msgChannel, err := i.mqClient.GetInputQueue(i.config.MqInputQueue)
	if err != nil {
		log.Panicf("%v", err)
	}

	i.wgroup.Add(1)

	go func() {
		defer i.wgroup.Done()

		for {
			select {
			case <-ctx.Done():
				i.mqClient.Close()
				return
			case msg := <-msgChannel:
				inputCh <- &InputMessageMQ{message: msg}
			}
		}
	}()
}

func (i *IOConnectorMQ) SetInputChannel(ctx context.Context, inputCh chan InputMessage) {
	i.mqInputOnce.Do(func() { i.InputInit(ctx, inputCh) })
}

func (i *IOConnectorMQ) OutputInit(ctx context.Context, outputCh chan string) {
	err := i.mqClient.CheckOutputQueue(i.config.MqOutputExchange, i.config.MqOutputRoutingKey)
	if err != nil {
		log.Fatalf("cannot connect to output queue, %v", err)
	}

	i.wgroup.Add(1)

	go func() {
		defer i.wgroup.Done()

		for {
			select {
			case <-ctx.Done():
				// close output connection
				return
			case output := <-outputCh:
				err = i.mqClient.Publish(i.config.MqOutputExchange, i.config.MqOutputRoutingKey, output)
				if err != nil {
					log.Printf("publish error, %v", err)
				}
				log.Printf("output: %s\n", output)
			}
		}
	}()
}

func (i *IOConnectorMQ) SetOutputChannel(ctx context.Context, outputCh chan string) {
	i.mqOutputOnce.Do(func() { i.OutputInit(ctx, outputCh) })
}

func NewIOConnectorMQ(config *ConfigRunner, wgroup *sync.WaitGroup) (*IOConnectorMQ, error) {
	mqClient, err := NewMQClient(config.MqUrl)
	if err != nil {
		return nil, err
	}
	return &IOConnectorMQ{
		mqInputOnce:  sync.Once{},
		mqOutputOnce: sync.Once{},
		config:       config,
		mqClient:     mqClient,
		wgroup:       wgroup,
	}, nil
}
