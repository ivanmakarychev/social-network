package tape

import (
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/ivanmakarychev/social-network/internal/config"
	"github.com/ivanmakarychev/social-network/internal/models"
	"github.com/ivanmakarychev/social-network/internal/repository"
)

type (
	Provider interface {
		GetTape(profileID models.ProfileID) ([]*models.Update, error)
		SubscribeOnUpdates(profileID models.ProfileID) (<-chan *models.Update, error)
		UnsubscribeFromUpdates(profileID models.ProfileID) error
	}

	CachingProvider struct {
		cache         *ristretto.Cache
		repo          repository.UpdatesRepo
		queue         BroadcastQueue
		limit         int
		subscriptions UpdatesSubscriptionManager
	}

	UpdatesSubscriptionManager interface {
		Subscribe(id models.ProfileID) (<-chan *models.Update, error)
		Unsubscribe(id models.ProfileID) error
	}

	UpdatesSubscriptionManagerImpl struct {
		lock          sync.RWMutex
		subscriptions map[models.ProfileID]chan *models.Update
		queue         DirectQueue
	}
)

func (p *CachingProvider) SubscribeOnUpdates(profileID models.ProfileID) (<-chan *models.Update, error) {
	return p.subscriptions.Subscribe(profileID)
}

func (p *CachingProvider) UnsubscribeFromUpdates(profileID models.ProfileID) error {
	return p.subscriptions.Unsubscribe(profileID)
}

func NewCachingProvider(
	cfg config.Updates,
	repo repository.UpdatesRepo,
	queue BroadcastQueue,
	subscriptionManager UpdatesSubscriptionManager,
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
		cache:         cache,
		limit:         cfg.Limit,
		repo:          repo,
		queue:         queue,
		subscriptions: subscriptionManager,
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

func (u *UpdatesSubscriptionManagerImpl) Subscribe(id models.ProfileID) (<-chan *models.Update, error) {
	u.lock.Lock()
	defer u.lock.Unlock()
	if ch, ok := u.subscriptions[id]; ok {
		return ch, nil
	}
	ch := make(chan *models.Update, 16)
	u.subscriptions[id] = ch
	u.queue.Subscribe(id, u.update)
	return ch, nil
}

func (u *UpdatesSubscriptionManagerImpl) Unsubscribe(id models.ProfileID) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.queue.Unsubscribe(id)
	if ch, ok := u.subscriptions[id]; ok {
		close(ch)
		delete(u.subscriptions, id)
	}
	return nil
}

func (u *UpdatesSubscriptionManagerImpl) update(upd UpdateWithSubscriber) {
	u.lock.RLock()
	defer u.lock.RUnlock()
	if ch, ok := u.subscriptions[upd.Subscriber]; ok {
		ch <- upd.Update
	}
}

func NewUpdatesSubscriptionManagerImpl(q DirectQueue) *UpdatesSubscriptionManagerImpl {
	return &UpdatesSubscriptionManagerImpl{
		subscriptions: map[models.ProfileID]chan *models.Update{},
		queue:         q,
	}
}
