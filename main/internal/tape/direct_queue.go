package tape

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/ivanmakarychev/social-network/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type (
	DirectQueue interface {
		Add(ctx context.Context, u UpdateWithSubscriber) error
		Subscribe(profileID models.ProfileID, callback UpdateSubscriptionCallback)
		Unsubscribe(profileID models.ProfileID)
	}

	DirectQueueImpl struct {
		connStr       string
		conn          *amqp.Connection
		ch            *amqp.Channel
		q             amqp.Queue
		msgs          <-chan amqp.Delivery
		subscriptions map[models.ProfileID]UpdateSubscriptionCallback
		lock          sync.RWMutex
	}
)

func (q *DirectQueueImpl) Add(ctx context.Context, u UpdateWithSubscriber) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	err = q.ch.PublishWithContext(
		ctx,
		directUpdatesExchangeName,
		u.Subscriber.String(),
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		})
	return err
}

func (q *DirectQueueImpl) Subscribe(profileID models.ProfileID, callback UpdateSubscriptionCallback) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if _, ok := q.subscriptions[profileID]; ok {
		return
	}
	q.subscriptions[profileID] = callback
	err := q.ch.QueueBind(
		q.q.Name,
		profileID.String(),
		directUpdatesExchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Println("failed to bind queue for direct updates", err)
	}
}

func (q *DirectQueueImpl) Unsubscribe(profileID models.ProfileID) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if _, ok := q.subscriptions[profileID]; !ok {
		return
	}
	delete(q.subscriptions, profileID)
	err := q.ch.QueueUnbind(
		q.q.Name,
		profileID.String(),
		directUpdatesQueueName,
		nil,
	)
	if err != nil {
		log.Println("failed to unbind queue for direct updates", err)
	}
}

func (q *DirectQueueImpl) Init() error {
	var err error
	q.conn, err = amqp.Dial(q.connStr)
	if err != nil {
		return err
	}
	q.ch, err = q.conn.Channel()
	if err != nil {
		return err
	}
	err = q.ch.ExchangeDeclare(
		directUpdatesExchangeName,
		"direct",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	q.q, err = q.ch.QueueDeclare(
		directUpdatesQueueName,
		false,
		false,
		true,
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
	go func() {
		for d := range q.msgs {
			q.consume(d)
		}
	}()
	return err
}

func (q *DirectQueueImpl) Close() {
	if q.conn != nil {
		_ = q.conn.Close()
	}
	if q.ch != nil {
		_ = q.ch.Close()
	}
}

func NewDirectQueue(connStr string) *DirectQueueImpl {
	return &DirectQueueImpl{
		connStr:       connStr,
		subscriptions: map[models.ProfileID]UpdateSubscriptionCallback{},
	}
}

func (q *DirectQueueImpl) consume(d amqp.Delivery) {
	u := UpdateWithSubscriber{}
	err := json.Unmarshal(d.Body, &u)
	if err != nil {
		log.Println("direct updates queue failed to unmarshal delivery:", err)
		return
	}
	q.getCallback(u.Subscriber)(u)
}

func (q *DirectQueueImpl) getCallback(subscriber models.ProfileID) UpdateSubscriptionCallback {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if callback, ok := q.subscriptions[subscriber]; ok {
		return callback
	}
	return emptyUpdateHandler
}

func emptyUpdateHandler(_ UpdateWithSubscriber) {}

const (
	directUpdatesExchangeName = "updates_direct"
	directUpdatesQueueName    = "updates_direct_queue"
)
