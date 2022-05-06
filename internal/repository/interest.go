package repository

import (
	"database/sql"

	"github.com/ivanmakarychev/social-network/internal/models"
)

type InterestsRepository interface {
	GetInterests() ([]models.Interest, error)
}

type InterestsRepositoryImpl struct {
	db *sql.DB
}

func NewInterestsRepositoryImpl(db *sql.DB) *InterestsRepositoryImpl {
	return &InterestsRepositoryImpl{db: db}
}

func (i *InterestsRepositoryImpl) GetInterests() ([]models.Interest, error) {
	var interests []models.Interest
	rows, err := i.db.Query("select interest_id, name from interests")
	if err != nil {
		return interests, err
	}
	defer rows.Close()
	for rows.Next() {
		interest := models.Interest{}
		err = rows.Scan(&interest.ID, &interest.Name)
		if err != nil {
			return interests, err
		}
		interests = append(interests, interest)
	}
	return interests, nil
}
