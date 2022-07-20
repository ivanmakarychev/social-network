package tape

import (
	"github.com/ivanmakarychev/social-network/internal/models"
	"github.com/ivanmakarychev/social-network/internal/repository"
)
import "github.com/dgraph-io/ristretto"

type (
	Provider interface {
		GetTape(profileID models.ProfileID) ([]*models.Update, error)
	}

	CachingProvider struct {
		cache *ristretto.Cache
	}

	PersistentProvider struct {
		repo repository.UpdatesRepo
	}
)

func NewPersistentProvider(repo repository.UpdatesRepo) *PersistentProvider {
	return &PersistentProvider{repo: repo}
}

func (p *PersistentProvider) GetTape(profileID models.ProfileID) ([]*models.Update, error) {
	return p.repo.GetUpdates(profileID)
}
