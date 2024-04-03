package main

// @Title        ioconnectormq.go
// @Description  实现基于MQ队列的输入输出方式（用于正式测量）
// @Create       dlchang (2024/04/03 15:00)

// todo:
// 1. prefetch & ack
// 2. publish & consume 使用两个connection
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
		inputQueue,
		"",
		true,
		false,
		false,
		false,
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

type IOConnectorMQ struct {
	mqInputOnce  sync.Once
	mqOutputOnce sync.Once
	wgroup       *sync.WaitGroup
	config       *ConfigRunner
	mqClient     *MQClient
	inputCh      chan string
	outputCh     chan string
}

func (i *IOConnectorMQ) InputInit(ctx context.Context) {
	err := i.mqClient.CheckInputQueue(i.config.MqInputQueue)
	if err != nil {
		log.Panicf("%v", err)
	}
	msgChannel, err := i.mqClient.GetInputQueue(i.config.MqInputQueue)
	if err != nil {
		log.Panicf("%v", err)
	}

	i.inputCh = make(chan string)
	i.wgroup.Add(1)

	go func() {
		defer i.wgroup.Done()
		for {
			select {
			case <-ctx.Done():
				i.mqClient.Close()
				return
			case msg := <-msgChannel:
				input := string(msg.Body)
				i.inputCh <- input
			}
		}
	}()
}

func (i *IOConnectorMQ) InputChannel(ctx context.Context) chan string {
	i.mqInputOnce.Do(func() { i.InputInit(ctx) })
	return i.inputCh
}

func (i *IOConnectorMQ) OutputInit(ctx context.Context) {
	err := i.mqClient.CheckOutputQueue(i.config.MqOutputExchange, i.config.MqOutputRoutingKey)
	if err != nil {
		log.Fatalf("cannot connect to output queue, %v", err)
	}

	i.outputCh = make(chan string)
	i.wgroup.Add(1)

	go func() {
		defer i.wgroup.Done()

		for {
			select {
			case <-ctx.Done():
				// close output connection
				return
			case output := <-i.outputCh:
				err = i.mqClient.Publish(i.config.MqOutputExchange, i.config.MqOutputRoutingKey, output)
				if err != nil {
					log.Printf("publish error, %v", err)
				}
				log.Printf("output: %s\n", output)
			}
		}
	}()
}

func (i *IOConnectorMQ) OutputChannel(ctx context.Context) chan string {
	i.mqOutputOnce.Do(func() { i.OutputInit(ctx) })
	return i.outputCh
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
