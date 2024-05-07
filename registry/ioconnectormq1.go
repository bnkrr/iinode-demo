package main

// @Title        ioconnectormq1.go
// @Description  统一的模式MQ队列输入输出（用于正式测量）
// @Create       dlchang (2024/04/18 18:00)

// 这个MQ使用统一的方式输入和输出到给定队列，实现免配置使用
// 利用两个connection分隔输入输出，建立多个channel和多个队列复用connection
import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MQChannel struct {
	channel *amqp.Channel
}

func (c *MQChannel) Close() error {
	return c.channel.Close()
}

func (c *MQChannel) QueueDeclare(exchange string, routingKey string) (string, error) {
	// 声明队列
	if exchange == "" {
		q, err := c.channel.QueueDeclare(
			routingKey, // name
			true,       // durable
			false,      // delete when unused
			false,      // exclusive
			false,      // no-wait
			nil,        // arguments
		)
		if err != nil {
			return "", err
		}
		return q.Name, nil
	}

	// 声明exchange
	err := c.channel.ExchangeDeclarePassive(
		exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return "", err
	}
	return exchange, nil
}

func (c *MQChannel) QueueBind(queue string, exchange string, routingKey string) error {
	if exchange == "" {
		return errors.New("cannot bind queue to another queue")
	}
	_, err := c.QueueDeclare("", queue)
	if err != nil {
		return err
	}
	return c.channel.QueueBind(queue, routingKey, exchange, false, nil)
}

func (c *MQChannel) GetInputQueue(inputQueueName string) (<-chan amqp.Delivery, error) {
	queueName, err := c.QueueDeclare("", inputQueueName)
	if err != nil {
		return nil, err
	}

	return c.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // autoAck
		false,     // exclusive
		false,     // noLocal
		false,     // noWait
		nil,
	)
}

func (c *MQChannel) Publish(outputExchange string, outputRoutingKey string, output string) error {
	err := c.channel.PublishWithContext(context.Background(),
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

func NewMQChannel(conn *amqp.Connection) (*MQChannel, error) {
	channel, err := conn.Channel() // 可并发执行
	if err != nil {
		return nil, err
	}

	return &MQChannel{
		channel: channel,
	}, nil

}

// manage all
type MQConnectorManager struct {
	url                    string
	outputExhange          string
	inputQueuePrefix       string
	outputRoutingKeyPrefix string
	inputConn              *amqp.Connection
	outputConn             *amqp.Connection
}

func (mg *MQConnectorManager) Close() error {
	err1 := mg.inputConn.Close()
	err2 := mg.outputConn.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func (mg *MQConnectorManager) GetInputQueueName(serviceName string) string {
	return mg.inputQueuePrefix + strings.ToLower(serviceName)
}

func (mg *MQConnectorManager) GetOutputRoutingKey(serviceName string) string {
	return mg.outputRoutingKeyPrefix + strings.ToLower(serviceName)
}

func (mg *MQConnectorManager) GetOutputQueueName(nextServiceName string) string {
	return mg.GetInputQueueName(nextServiceName)
}

func (mg *MQConnectorManager) GetNewInputConnector(c *ConfigRunner, wgroup *sync.WaitGroup) (*InputConnectorMQ1, error) {
	channel, err := NewMQChannel(mg.inputConn)
	if err != nil {
		return nil, err
	}

	return &InputConnectorMQ1{
		once:           sync.Once{},
		wgroup:         wgroup,
		serviceName:    c.Name,
		inputQueueName: mg.GetInputQueueName(c.Name),
		mqChannel:      channel,
	}, nil
}

func (mg *MQConnectorManager) GetNewOutputConnector(c *ConfigRunner, wgroup *sync.WaitGroup) (*OutputConnectorMQ1, error) {
	channel, err := NewMQChannel(mg.outputConn)
	if err != nil {
		return nil, err
	}

	outputQueue := ""
	if c.Mq1NextService != "" {
		outputQueue = mg.GetOutputQueueName(c.Mq1NextService)
	} else if c.Mq1OutputQueue != "" {
		outputQueue = c.Mq1OutputQueue
	}

	return &OutputConnectorMQ1{
		once:             sync.Once{},
		wgroup:           wgroup,
		serviceName:      c.Name,
		outputExhange:    mg.outputExhange,
		outputRoutingKey: mg.GetOutputRoutingKey(c.Name),
		outputQueue:      outputQueue,
		mqChannel:        channel,
	}, nil
}

func NewMQConnectorManger(config ConfigMQ) (*MQConnectorManager, error) {
	inputConn, err := amqp.Dial(config.Url)
	if err != nil {
		return nil, err
	}
	outputConn, err := amqp.Dial(config.Url)
	if err != nil {
		return nil, err
	}
	return &MQConnectorManager{
		url:                    config.Url,
		outputExhange:          config.OutputExchange,
		inputQueuePrefix:       config.InputQueuePrefix,
		outputRoutingKeyPrefix: config.OutputRoutingKeyPrefix,
		inputConn:              inputConn,
		outputConn:             outputConn,
	}, nil
}

// input
type InputConnectorMQ1 struct {
	once           sync.Once
	wgroup         *sync.WaitGroup
	serviceName    string
	inputQueueName string
	mqChannel      *MQChannel
}

func (i *InputConnectorMQ1) PrepareAndGetInputQueue() (<-chan amqp.Delivery, error) {
	return i.mqChannel.GetInputQueue(i.inputQueueName)
}

func (i *InputConnectorMQ1) Init(ctx context.Context, inputCh chan InputMessage) {
	msgChannel, err := i.PrepareAndGetInputQueue()
	if err != nil {
		log.Panicf("%v", err)
	}

	i.wgroup.Add(1)

	go func() {
		defer i.wgroup.Done()

		for {
			select {
			case <-ctx.Done():
				i.mqChannel.Close()
				return
			case msg := <-msgChannel:
				inputCh <- &InputMessageMQ{message: msg}
			}
		}
	}()
}

func (i *InputConnectorMQ1) SetInputChannel(ctx context.Context, inputCh chan InputMessage) {
	i.once.Do(func() { i.Init(ctx, inputCh) })
}

// output
type OutputConnectorMQ1 struct {
	once             sync.Once
	wgroup           *sync.WaitGroup
	serviceName      string
	outputExhange    string
	outputRoutingKey string
	outputQueue      string
	mqChannel        *MQChannel
}

func (o *OutputConnectorMQ1) PrepareOutputQueue() error {
	_, err := o.mqChannel.QueueDeclare(o.outputExhange, o.outputRoutingKey)
	if err != nil {
		return err
	}
	if o.outputQueue != "" {
		return o.mqChannel.QueueBind(o.outputQueue, o.outputExhange, o.outputRoutingKey)
	}
	return err
}

func (o *OutputConnectorMQ1) Publish(message string) error {
	return o.mqChannel.Publish(o.outputExhange, o.outputRoutingKey, message)
}

func (o *OutputConnectorMQ1) Init(ctx context.Context, outputCh chan string) {
	err := o.PrepareOutputQueue()
	if err != nil {
		log.Fatalf("cannot connect to output queue, %v", err)
	}

	o.wgroup.Add(1)

	go func() {
		defer o.wgroup.Done()

		for {
			select {
			case <-ctx.Done():
				// close output connection
				o.mqChannel.Close()
				return
			case output := <-outputCh:
				err = o.Publish(output)
				if err != nil {
					log.Printf("publish error, %v", err)
				}
				log.Printf("output: %s\n", output)
			}
		}
	}()
}

func (o *OutputConnectorMQ1) SetOutputChannel(ctx context.Context, outputCh chan string) {
	o.once.Do(func() { o.Init(ctx, outputCh) })
}
