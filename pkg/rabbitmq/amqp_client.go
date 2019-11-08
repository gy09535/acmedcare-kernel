package rabbitmq

import (
	"acmed.com/kernel/pkg/rpc"
	"bytes"
	"errors"
	"github.com/streadway/amqp"
	"strings"
	"time"
)

type AmqpClient struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	topics         string
	nodes          string
	receiveChannel chan int
}

func (amqpClient *AmqpClient) Init(address string) (err error) {

	amqpClient.Close()
	amqpClient.conn, err = amqp.Dial(address)
	if err != nil {
		return err
	}

	amqpClient.channel, err = amqpClient.conn.Channel()
	amqpClient.receiveChannel = make(chan int)
	if err != nil {
		return err
	}

	return nil
}

func (amqpClient *AmqpClient) Ping() (err error) {

	if amqpClient.channel == nil {
		return errors.New("rabbitmq is not initialize")
	}

	err = amqpClient.channel.ExchangeDeclare("ping.ping", "topic", false, true, false, true, nil)
	if err != nil {
		return err
	}

	msgContent := "ping.ping"
	err = amqpClient.channel.Publish("ping.ping", "ping.ping", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msgContent),
	})

	if err != nil {
		return err
	}

	err = amqpClient.channel.ExchangeDelete("ping.ping", false, false)
	return err
}

func (amqpClient *AmqpClient) Publish(topic, msg string) (err error) {

	if amqpClient.topics == "" || !strings.Contains(amqpClient.topics, topic) {
		err = amqpClient.channel.ExchangeDeclare(topic, "topic", true, false, false, true, nil)
		if err != nil {
			return err
		}

		amqpClient.topics += "  " + topic + "  "
	}

	err = amqpClient.channel.Publish(topic, topic, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})

	return nil
}

func (amqpClient *AmqpClient) Receive(topic, route string, queue string, reader func(msg *string)) {

	if amqpClient.topics == "" || !strings.Contains(amqpClient.topics, topic) {
		err := amqpClient.channel.ExchangeDeclare(topic, "topic", true, false, false, true, nil)
		if err != nil {
			rpc.Logger.Println("receive msg from rabbitmq server error:" + err.Error())
		}

		amqpClient.topics += "  " + topic + "  "
	}

	if amqpClient.nodes == "" || !strings.Contains(amqpClient.nodes, queue) {
		_, err := amqpClient.channel.QueueDeclare(queue, false, true, false, true, nil)
		if err != nil {
			rpc.Logger.Println("receive msg from rabbitmq server error:" + err.Error())
		}

		err = amqpClient.channel.QueueBind(queue, route, topic, true, nil)
		if err != nil {
			rpc.Logger.Println("receive msg from rabbitmq server error:" + err.Error())
		}

		amqpClient.nodes += "  " + queue + "  "
	}

	go func() {
		for {
			select {
			case <-time.After(time.Second * 10):
				messages, err := amqpClient.channel.Consume(queue, "", true, false, false, false, nil)
				if err != nil {
					rpc.Logger.Printf("receive msg from control center error,%s\r\n", err.Error())
				}

				for d := range messages {
					s := bytesToString(&(d.Body))
					reader(s)
				}
				break
			case <-amqpClient.receiveChannel:
				return
			}
		}
	}()
}

func (amqpClient *AmqpClient) Close() {

	if amqpClient.channel != nil {
		amqpClient.channel.Close()
		amqpClient.channel = nil
	}

	if amqpClient.conn != nil {
		amqpClient.conn.Close()
		amqpClient.conn = nil
	}

	if amqpClient.receiveChannel != nil {
		amqpClient.receiveChannel <- 1
	}
}

func bytesToString(b *[]byte) *string {
	s := bytes.NewBuffer(*b)
	r := s.String()
	return &r
}
