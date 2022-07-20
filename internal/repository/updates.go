package repository

import (
	"github.com/ivanmakarychev/social-network/internal/models"
)

type (
	UpdatesRepo interface {
		SaveUpdate(u *models.Update) error
		GetUpdates(subscriberID models.ProfileID) ([]*models.Update, error)
		SaveSubscription(rq models.SubscriptionRq) error
	}

	ClusterUpdatesRepo struct {
		db Cluster
	}
)

func NewClusterUpdatesRepo(db Cluster) *ClusterUpdatesRepo {
	return &ClusterUpdatesRepo{db: db}
}

func (r *ClusterUpdatesRepo) SaveUpdate(u *models.Update) error {
	//todo impl
	return nil
}

func (r *ClusterUpdatesRepo) GetUpdates(subscriberID models.ProfileID) ([]*models.Update, error) {
	//todo impl
	return []*models.Update{
		{
			ID: 1,
			Author: models.ProfileMain{
				Name:    "Иван",
				Surname: "Петров",
			},
			DateFmt: "10:25 11.07.2022",
			Text:    "Я сегодня поел картошки.",
		},
		{
			ID: 2,
			Author: models.ProfileMain{
				Name:    "Петр",
				Surname: "Иванов",
			},
			DateFmt: "06:08 20.07.2022",
			Text:    "Попытки изолировать нашу страну обречены на провал.",
		},
	}, nil
}

func (r *ClusterUpdatesRepo) SaveSubscription(rq models.SubscriptionRq) error {
	//todo impl
	return nil
}
