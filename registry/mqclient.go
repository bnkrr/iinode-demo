package main

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MQClient struct {
	url     string
	conn    *amqp.Connection
	channel *amqp.Channel
}

func (m *MQClient) GetInputQueue(inputQueue string) (<-chan amqp.Delivery, error) {
	q, err := m.channel.QueueDeclarePassive(
		inputQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	msgChannel, err := m.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return msgChannel, nil
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
