package tape

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ivanmakarychev/social-network/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type (
	UpdateWithSubscriber struct {
		Subscriber models.ProfileID `json:"subscriber"`
		Update     *models.Update   `json:"update"`
	}

	UpdatesRouter interface {
		Add(ctx context.Context, u UpdateWithSubscriber) error
	}

	BroadcastQueue interface {
		Add(ctx context.Context, u UpdateWithSubscriber) error
		Subscribe(UpdateSubscriptionCallback)
	}

	BroadcastQueueImpl struct {
		connStr       string
		conn          *amqp.Connection
		ch            *amqp.Channel
		q             amqp.Queue
		msgs          <-chan amqp.Delivery
		subscriptions []UpdateSubscriptionCallback
	}

	UpdateSubscriptionCallback func(u UpdateWithSubscriber)
)

func NewBroadcastQueue(connStr string) *BroadcastQueueImpl {
	return &BroadcastQueueImpl{connStr: connStr}
}

func (q *BroadcastQueueImpl) Add(ctx context.Context, u UpdateWithSubscriber) error {
	body, err := json.Marshal(u)
	if err != nil {
		return err
	}
	err = q.ch.PublishWithContext(
		ctx,
		"",
		q.q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	return err
}

func (q *BroadcastQueueImpl) Subscribe(f UpdateSubscriptionCallback) {
	q.subscriptions = append(q.subscriptions, f)
}

func (q *BroadcastQueueImpl) Init() error {
	var err error
	q.conn, err = amqp.Dial(q.connStr)
	if err != nil {
		return err
	}
	q.ch, err = q.conn.Channel()
	if err != nil {
		return err
	}
	q.q, err = q.ch.QueueDeclare(
		"updates",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	q.msgs, err = q.ch.Consume(
		q.q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	go func() {
		for d := range q.msgs {
			q.consume(d)
		}
	}()
	return err
}

func (q *BroadcastQueueImpl) Close() {
	if q.conn != nil {
		_ = q.conn.Close()
	}
	if q.ch != nil {
		_ = q.ch.Close()
	}
}

func (q *BroadcastQueueImpl) consume(d amqp.Delivery) {
	u := UpdateWithSubscriber{}
	err := json.Unmarshal(d.Body, &u)
	if err != nil {
		log.Println("updates queue failed to unmarshal delivery:", err)
		return
	}
	for _, callback := range q.subscriptions {
		callback(u)
	}
}
