package repository

import (
	"github.com/ivanmakarychev/social-network/internal/models"
)

type CitiesRepository interface {
	GetCities() ([]models.City, error)
}

type CitiesRepositoryImpl struct {
	db Cluster
}

func NewCitiesRepositoryImpl(db Cluster) *CitiesRepositoryImpl {
	return &CitiesRepositoryImpl{db: db}
}

func (c *CitiesRepositoryImpl) GetCities() ([]models.City, error) {
	var cities []models.City
	rows, err := c.db.Master().Query("select city_id, name from cities where city_id <> 0")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			city := models.City{}
			err = rows.Scan(&city.ID, &city.Name)
			if err != nil {
				break
			}
			cities = append(cities, city)
		}
	}
	return cities, err
}
