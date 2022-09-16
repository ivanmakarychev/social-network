package repository

import (
	"github.com/Masterminds/squirrel"
	"github.com/ivanmakarychev/social-network/internal/models"
)

type (
	UpdatesRepo interface {
		SaveUpdate(u *models.Update) error
		GetUpdates(rq GetUpdatesRq) ([]*models.Update, error)
		SaveSubscription(rq models.SubscriptionRq) error
		GetSubscribers(publisher models.ProfileID) ([]models.ProfileID, error)
	}

	ClusterUpdatesRepo struct {
		db Cluster
	}

	GetUpdatesRq struct {
		SubscriberID models.ProfileID
		Limit        uint64
	}
)

func NewClusterUpdatesRepo(db Cluster) *ClusterUpdatesRepo {
	return &ClusterUpdatesRepo{db: db}
}

func (r *ClusterUpdatesRepo) SaveUpdate(u *models.Update) error {
	query, args, err := squirrel.Insert("updates").
		Columns("publisher_id", "text").
		Values(u.Author.ID, u.Text).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Master().Exec(query, args...)
	return err
}

func (r *ClusterUpdatesRepo) GetUpdates(rq GetUpdatesRq) ([]*models.Update, error) {
	query, args, err := squirrel.Select("u.update_id", "u.publisher_id", "u.ts", "u.text", "p.first_name", "p.surname").
		From("updates u").
		Join("subscriptions s on s.publisher_id = u.publisher_id").
		Join("profile p on p.profile_id = u.publisher_id").
		Where(squirrel.Eq{"s.subscriber_id": rq.SubscriberID}).
		OrderBy("u.update_id desc").
		Limit(rq.Limit).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Replica().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*models.Update
	for rows.Next() {
		u := &models.Update{}
		err = rows.Scan(&u.ID, &u.Author.ID, &u.TS, &u.Text, &u.Author.Name, &u.Author.Surname)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, nil
}

func (r *ClusterUpdatesRepo) SaveSubscription(rq models.SubscriptionRq) error {
	query, args, err := squirrel.Insert("subscriptions").
		Columns("subscriber_id", "publisher_id").
		Values(rq.SubscriberID, rq.PublisherID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Master().Exec(query, args...)
	return err
}

func (r *ClusterUpdatesRepo) GetSubscribers(publisher models.ProfileID) ([]models.ProfileID, error) {
	query, args, err := squirrel.Select("subscriber_id").
		From("subscriptions").
		Where(squirrel.Eq{"publisher_id": publisher}).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Replica().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rs []models.ProfileID
	for rows.Next() {
		var p models.ProfileID
		err = rows.Scan(&p)
		if err != nil {
			return nil, err
		}
		rs = append(rs, p)
	}
	return rs, nil
}
