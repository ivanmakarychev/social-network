package tape

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/ivanmakarychev/social-network/internal/config"
	"github.com/ivanmakarychev/social-network/internal/models"
	"github.com/ivanmakarychev/social-network/internal/repository"
)

type (
	Publisher interface {
		Publish(u *models.Update) error
	}

	QueuePublisher struct {
		repo  repository.UpdatesRepo
		queue Queue
		cfg   config.Updates
	}
)

func NewQueuePublisher(
	repo repository.UpdatesRepo,
	q Queue,
	cfg config.Updates,
) *QueuePublisher {
	rand.Seed(time.Now().Unix())
	return &QueuePublisher{
		repo:  repo,
		queue: q,
		cfg:   cfg,
	}
}

func (p *QueuePublisher) Publish(u *models.Update) error {
	err := p.repo.SaveUpdate(u)
	if err != nil {
		return err
	}
	s, err := p.repo.GetSubscribers(u.Author.ID)
	if err != nil {
		return err
	}
	if len(s) == 0 {
		return nil
	}
	rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
	numberOfReceivers := int(math.Ceil(float64(len(s)) * p.cfg.SubscribersFraction))
	for _, r := range s[:numberOfReceivers] {
		err = p.queue.Add(context.Background(), UpdateWithSubscriber{
			Subscriber: r,
			Update:     u,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
