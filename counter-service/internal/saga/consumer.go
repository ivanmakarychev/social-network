package saga

import (
	"encoding/json"
	"log"

	"github.com/ivanmakarychev/social-network/counter-service/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type (
	ConsumerImpl struct {
		connStr string
		conn    *amqp.Connection
		ch      *amqp.Channel
		q       amqp.Queue
		msgs    <-chan amqp.Delivery
	}
)

func NewConsumerImpl(connStr string) *ConsumerImpl {
	return &ConsumerImpl{connStr: connStr}
}

const (
	requestsQueueName = "message_counter_requests"
)

func (c *ConsumerImpl) Init() (<-chan *models.AlterCounterRequest, error) {
	var err error
	c.conn, err = amqp.Dial(c.connStr)
	if err != nil {
		return nil, err
	}
	c.ch, err = c.conn.Channel()
	if err != nil {
		return nil, err
	}
	c.q, err = c.ch.QueueDeclare(
		requestsQueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	c.msgs, err = c.ch.Consume(
		c.q.Name,
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
	ch := make(chan *models.AlterCounterRequest, 2)
	go func() {
		for d := range c.msgs {
			c.consume(d, ch)
		}
	}()
	return ch, err
}

func (c *ConsumerImpl) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	if c.ch != nil {
		_ = c.ch.Close()
	}
}

func (c *ConsumerImpl) consume(d amqp.Delivery, ch chan<- *models.AlterCounterRequest) {
	u := &models.AlterCounterRequest{}
	err := json.Unmarshal(d.Body, u)
	if err != nil {
		log.Println("failed to unmarshal alter counter requests:", err)
		return
	}
	ch <- u
}
