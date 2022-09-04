package saga

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ivanmakarychev/social-network/counter-service/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type (
	PublisherImpl struct {
		connStr     string
		conn        *amqp.Connection
		ch          *amqp.Channel
		q           amqp.Queue
		resultsChan <-chan *models.AlterCounterResult
	}
)

func NewPublisherImpl(connStr string) *PublisherImpl {
	return &PublisherImpl{connStr: connStr}
}

const (
	resultsQueueName = "message_counter_results"
)

func (p *PublisherImpl) Init() (chan<- *models.AlterCounterResult, error) {
	var err error
	p.conn, err = amqp.Dial(p.connStr)
	if err != nil {
		return nil, err
	}
	p.ch, err = p.conn.Channel()
	if err != nil {
		return nil, err
	}
	p.q, err = p.ch.QueueDeclare(
		resultsQueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	ch := make(chan *models.AlterCounterResult, 2)
	p.resultsChan = ch
	go func() {
		for r := range ch {
			p.publish(r)
		}
	}()
	return ch, nil
}

func (p *PublisherImpl) Close() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	if p.ch != nil {
		_ = p.ch.Close()
	}
}

func (p *PublisherImpl) publish(r *models.AlterCounterResult) {
	body, err := json.Marshal(r)
	if err != nil {
		log.Println("failed to marshal alter counter result", err.Error())
		return
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
		log.Println("failed to publish alter counter result", err.Error())
	}
}
