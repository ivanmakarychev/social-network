package repository

import (
	"fmt"
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/ivanmakarychev/social-network/internal/models"
	tarantool "github.com/tarantool/go-tarantool"
)

type FriendsRepo interface {
	GetFriends(profileID models.ProfileID) ([]models.Friend, error)
	GetFriendshipApplications(profileID models.ProfileID) ([]models.Friend, error)
	MakeFriend(owner, other models.ProfileID) error
	ConfirmFriendship(owner, other models.ProfileID) error
}

type FriendsRepoImpl struct {
	db Cluster
	t  *tarantool.Connection
}

func NewFriendsRepoImpl(db Cluster) *FriendsRepoImpl {
	return &FriendsRepoImpl{db: db}
}

func (f *FriendsRepoImpl) GetFriends(profileID models.ProfileID) ([]models.Friend, error) {
	selectFriends := squirrel.Select("p.profile_id", "first_name", "surname").
		From("friends f").
		Join(fmt.Sprintf("%s p on f.other_profile_id = p.profile_id", profileTableName)).
		Where(squirrel.Eq{"f.profile_id": profileID})
	query, args, err := selectFriends.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := f.db.Replica().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Friend
	for rows.Next() {
		friend := models.Friend{}
		err = rows.Scan(&friend.ID, &friend.Name, &friend.Surname)
		if err != nil {
			return nil, err
		}
		result = append(result, friend)
	}
	return result, nil
}

func (f *FriendsRepoImpl) GetFriendshipApplications(profileID models.ProfileID) ([]models.Friend, error) {
	selectFriends := squirrel.Select("p.profile_id", "first_name", "surname").
		From("friendship_application f").
		Join(fmt.Sprintf("%s p on f.profile_id = p.profile_id", profileTableName)).
		Where(squirrel.Eq{"f.other_profile_id": profileID})
	query, args, err := selectFriends.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := f.db.Master().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Friend
	for rows.Next() {
		friend := models.Friend{}
		err = rows.Scan(&friend.ID, &friend.Name, &friend.Surname)
		if err != nil {
			return nil, err
		}
		result = append(result, friend)
	}
	return result, nil
}

func (f *FriendsRepoImpl) MakeFriend(owner, other models.ProfileID) error {
	query, args, err := squirrel.Insert("friendship_application").
		Columns("profile_id", "other_profile_id").
		Values(owner, other).
		ToSql()
	_, err = f.db.Master().Exec(query, args...)
	return err
}

func (f *FriendsRepoImpl) ConfirmFriendship(owner, other models.ProfileID) error {
	var err error
	tx, err := f.db.Master().Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			errRollback := tx.Rollback()
			if errRollback != nil {
				log.Println("[confirm friendship] failed to rollback transaction: ", errRollback)
			}
		}
	}()

	query, args, err := squirrel.Insert("friends").
		Columns("profile_id", "other_profile_id").
		Values(owner, other).
		Values(other, owner).
		ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}

	query, args, err = squirrel.Delete("friendship_application").
		Where(squirrel.Or{
			squirrel.And{
				squirrel.Eq{"profile_id": owner},
				squirrel.Eq{"other_profile_id": other},
			},
			squirrel.And{
				squirrel.Eq{"profile_id": other},
				squirrel.Eq{"other_profile_id": owner},
			},
		}).ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, args...)
	return err
}
