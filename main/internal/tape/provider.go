package tape

import (
	"log"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/ivanmakarychev/social-network/internal/config"
	"github.com/ivanmakarychev/social-network/internal/models"
	"github.com/ivanmakarychev/social-network/internal/repository"
)

type (
	Provider interface {
		GetTape(profileID models.ProfileID) ([]*models.Update, error)
	}

	CachingProvider struct {
		cache *ristretto.Cache
		repo  repository.UpdatesRepo
		queue Queue
		limit int
	}
)

func NewCachingProvider(
	cfg config.Updates,
	repo repository.UpdatesRepo,
	queue Queue,
) *CachingProvider {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10_000_000,
		MaxCost:     100_000_000,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatalln("failed to init updates cache", err)
	}
	p := &CachingProvider{
		cache: cache,
		limit: cfg.Limit,
		repo:  repo,
		queue: queue,
	}
	queue.Subscribe(p.subscription)
	return p
}

func (p *CachingProvider) GetTape(profileID models.ProfileID) ([]*models.Update, error) {
	if rs, ok := p.getFromCache(profileID); ok {
		return rs, nil
	}
	rs, err := p.repo.GetUpdates(repository.GetUpdatesRq{
		SubscriberID: profileID,
		Limit:        uint64(p.limit),
	})
	if err != nil {
		return nil, err
	}
	p.putUpdatesInCache(profileID, rs)
	return rs, nil
}

func (p *CachingProvider) getFromCache(profileID models.ProfileID) ([]*models.Update, bool) {
	val, ok := p.cache.Get(uint64(profileID))
	if !ok {
		return nil, false
	}
	result, ok := val.([]*models.Update)
	return result, ok
}

func (p *CachingProvider) putUpdatesInCache(profileID models.ProfileID, u []*models.Update) {
	_ = p.cache.SetWithTTL(uint64(profileID), u, 0, 3*time.Hour)
}

func (p *CachingProvider) putInCache(profileID models.ProfileID, u *models.Update) {
	updates, _ := p.getFromCache(profileID)

	updates = append([]*models.Update{u}, updates...)[:min(p.limit, len(updates)+1)]
	p.putUpdatesInCache(profileID, updates)
}

func (p *CachingProvider) subscription(u UpdateWithSubscriber) {
	p.putInCache(u.Subscriber, u.Update)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
