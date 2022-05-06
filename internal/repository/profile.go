package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/ivanmakarychev/social-network/internal/models"
)

type ProfileRepo interface {
	GetProfile(userID models.ProfileID) (models.Profile, error)
	CreateProfileID() (models.ProfileID, error)
	SaveProfile(profile models.Profile) error
}

type ProfileRepoImpl struct {
	db          *sql.DB
	friendsRepo FriendsRepo
}

func NewProfileRepoImpl(
	db *sql.DB,
	friendsRepo FriendsRepo,
) *ProfileRepoImpl {
	return &ProfileRepoImpl{
		db:          db,
		friendsRepo: friendsRepo,
	}
}

const (
	profileTableName            = "profile"
	profileToInterestsTableName = "profile_interests"
)

func (p *ProfileRepoImpl) CreateProfileID() (models.ProfileID, error) {
	var err error
	tx, err := p.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			errRollback := tx.Rollback()
			if errRollback != nil {
				log.Println("[create profile] failed to rollback transaction: ", errRollback)
			}
		}
	}()

	_, err = tx.Exec(`insert into profile (name) values ('')`)
	if err != nil {
		return 0, err
	}

	var profileID models.ProfileID
	err = tx.QueryRow(`select LAST_INSERT_ID()`).Scan(&profileID)
	return profileID, err
}

func (p *ProfileRepoImpl) SaveProfile(profile models.Profile) error {
	var err error
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			errRollback := tx.Rollback()
			if errRollback != nil {
				log.Println("[save profile] failed to rollback transaction: ", errRollback)
			}
		}
	}()
	b := squirrel.Update(profileTableName).
		SetMap(map[string]interface{}{
			"name":       profile.Name,
			"surname":    profile.Surname,
			"city_id":    profile.City.ID,
			"birth_date": profile.BirthDate,
		}).
		Where(squirrel.Eq{
			"profile_id": profile.ID,
		})
	query, args, err := b.ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}

	query, args, err = squirrel.Delete(profileToInterestsTableName).
		Where(squirrel.Eq{
			"profile_id": profile.ID,
		}).ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}

	insertInterests := squirrel.Insert(profileToInterestsTableName).
		Columns("profile_id", "interest_id")
	for _, interest := range profile.Interests {
		insertInterests = insertInterests.Values(profile.ID, interest.ID)
	}
	query, args, err = insertInterests.ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}

	return err
}

func (p *ProfileRepoImpl) GetProfile(userID models.ProfileID) (models.Profile, error) {
	profile := models.Profile{}
	selectProfile := squirrel.Select("p.name", "surname", "birth_date", "p.city_id", "c.name").
		From(fmt.Sprintf("%s p", profileTableName)).
		Join("cities c on p.city_id = c.city_id").
		Where(squirrel.Eq{"profile_id": userID})
	query, args, err := selectProfile.ToSql()
	if err != nil {
		return profile, err
	}
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return profile, err
	}
	defer rows.Close()
	for rows.Next() {
		profile.ID = userID
		err = rows.Scan(&profile.Name, &profile.Surname, &profile.BirthDate, &profile.City.ID, &profile.City.Name)
		if err != nil {
			return profile, err
		}
	}
	if profile.ID == 0 {
		return profile, errors.New("not found")
	}

	profile.Friends, err = p.friendsRepo.GetFriends(userID)
	if err != nil {
		return profile, err
	}

	profile.FriendshipApplications, err = p.friendsRepo.GetFriendshipApplications(userID)
	if err != nil {
		return profile, err
	}

	query, args, err = squirrel.Select("i.interest_id", "i.name").
		From("profile_interests p").
		Join("interests i using(interest_id)").
		Where(squirrel.Eq{"p.profile_id": userID}).
		ToSql()
	if err != nil {
		return profile, err
	}
	rowsInterests, err := p.db.Query(query, args...)
	if err != nil {
		return profile, err
	}
	defer rowsInterests.Close()
	for rowsInterests.Next() {
		interest := models.Interest{}
		err = rowsInterests.Scan(&interest.ID, &interest.Name)
		if err != nil {
			return profile, err
		}
		profile.Interests = append(profile.Interests, interest)
	}

	return profile, nil
}
