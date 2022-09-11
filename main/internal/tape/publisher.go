package tape

import (
	"context"
	"log"
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

	RouterPublisher struct {
		repo   repository.UpdatesRepo
		router UpdatesRouter
		cfg    config.Updates
	}

	CompositeUpdatesRouter struct {
		routers []UpdatesRouter
	}
)

func NewCompositeUpdatesRouter(routers ...UpdatesRouter) *CompositeUpdatesRouter {
	return &CompositeUpdatesRouter{routers: routers}
}

func (c *CompositeUpdatesRouter) Add(ctx context.Context, u UpdateWithSubscriber) error {
	for _, r := range c.routers {
		err := r.Add(ctx, u)
		if err != nil {
			log.Println("failed to add update to router", err)
		}
	}
	return nil
}

func NewRouterPublisher(
	repo repository.UpdatesRepo,
	r UpdatesRouter,
	cfg config.Updates,
) *RouterPublisher {
	rand.Seed(time.Now().Unix())
	return &RouterPublisher{
		repo:   repo,
		router: r,
		cfg:    cfg,
	}
}

func (p *RouterPublisher) Publish(u *models.Update) error {
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
		err = p.router.Add(context.Background(), UpdateWithSubscriber{
			Subscriber: r,
			Update:     u,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
