package saga

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type (
	PublisherImpl struct {
		connStr string
		conn    *amqp.Connection
		ch      *amqp.Channel
		q       amqp.Queue
	}
)

func NewPublisherImpl(connStr string) *PublisherImpl {
	return &PublisherImpl{connStr: connStr}
}

const (
	requestsQueueName = "message_counter_requests"
)

func (p *PublisherImpl) Init() error {
	var err error
	p.conn, err = amqp.Dial(p.connStr)
	if err != nil {
		return err
	}
	p.ch, err = p.conn.Channel()
	if err != nil {
		return err
	}
	p.q, err = p.ch.QueueDeclare(
		requestsQueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	return err
}

func (p *PublisherImpl) Close() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	if p.ch != nil {
		_ = p.ch.Close()
	}
}

func (p *PublisherImpl) Publish(r *models.AlterCounterRequest) error {
	body, err := json.Marshal(r)
	if err != nil {
		log.Println("failed to marshal alter counter request", err.Error())
		return err
	}
	err = p.ch.PublishWithContext(
		context.Background(),
		"",
		p.q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		log.Println("failed to publish alter counter request", err.Error())
	}
	return err
}
