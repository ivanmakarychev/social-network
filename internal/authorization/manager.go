package authorization

import (
	"errors"
	"log"
	"regexp"

	"github.com/Masterminds/squirrel"
	"github.com/ivanmakarychev/social-network/internal/models"
	"github.com/ivanmakarychev/social-network/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type LoginData struct {
	Login    string
	Password string
}

var (
	// ErrUnauthorized не авторизован
	ErrUnauthorized = errors.New("unauthorized")
	// ErrBadPassword пароль не соответствует требованиям
	ErrBadPassword = errors.New("bad password")
)

type Manager interface {
	GetUserID(login LoginData) (models.ProfileID, error)
	SaveLogin(profileID models.ProfileID, login LoginData) error
}

type ManagerImpl struct {
	db repository.Cluster
}

func NewManagerImpl(db repository.Cluster) *ManagerImpl {
	return &ManagerImpl{db: db}
}

func (m *ManagerImpl) SaveLogin(profileID models.ProfileID, login LoginData) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(login.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	query, args, err := squirrel.Insert("logins").
		Columns("login", "password_hash", "profile_id").
		Values(login.Login, hash, profileID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = m.db.Master().Exec(query, args...)
	return err
}

func (m *ManagerImpl) GetUserID(login LoginData) (models.ProfileID, error) {
	query, args, err := squirrel.Select("profile_id", "password_hash").
		From("logins").
		Where(squirrel.Eq{"login": login.Login}).
		ToSql()
	if err != nil {
		return 0, err
	}
	var profileID models.ProfileID
	var hashedPassword []byte
	rows, err := m.db.Master().Query(query, args...)
	if err != nil {
		log.Println("failed to get profile id by login: ", err)
		return 0, ErrUnauthorized
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&profileID, &hashedPassword)
		if err != nil {
			log.Println("failed to scan profile id from query result: ", err)
			return 0, ErrUnauthorized
		}
		break
	}
	if profileID == 0 {
		return 0, ErrUnauthorized
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(login.Password))
	if err != nil {
		log.Println("password check failed: ", err)
		return 0, ErrUnauthorized
	}
	return profileID, nil
}

func ValidatePassword(password string) error {
	const maxByteSize = 72
	if len([]byte(password)) > maxByteSize {
		return ErrBadPassword
	}
	if passwordRegexp.MatchString(password) {
		return nil
	}
	return ErrBadPassword
}

var (
	passwordRegexp = regexp.MustCompile("^[a-zA-Z0-9_\\-!@#&%$.,:;]{8,}$")
)
