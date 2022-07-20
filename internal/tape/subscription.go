package tape

import (
	"github.com/ivanmakarychev/social-network/internal/models"
	"github.com/ivanmakarychev/social-network/internal/repository"
)

type (
	Subscription interface {
		Subscribe(rq models.SubscriptionRq) error
		GetSomeSubscribers(id models.ProfileID) ([]models.ProfileID, error)
	}

	SubscriptionImpl struct {
		repo repository.UpdatesRepo
	}
)

func NewSubscriptionImpl(repo repository.UpdatesRepo) *SubscriptionImpl {
	return &SubscriptionImpl{repo: repo}
}

func (s *SubscriptionImpl) Subscribe(rq models.SubscriptionRq) error {
	return s.repo.SaveSubscription(rq)
}

func (s *SubscriptionImpl) GetSomeSubscribers(id models.ProfileID) ([]models.ProfileID, error) {
	panic("implement me")
}
