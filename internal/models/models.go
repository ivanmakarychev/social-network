package models

import "time"

// ProfileID id профиля
type ProfileID uint64

// ProfileMain основная информация о профиле пользователя
type ProfileMain struct {
	ID      ProfileID
	Name    string
	Surname string
}

// Profile профиль пользователя
type Profile struct {
	ProfileMain
	BirthDate              time.Time
	City                   City
	Friends                []Friend
	FriendshipApplications []Friend // заявки в друзья
	Interests              []Interest
}

// Friend данные о друге
type Friend struct {
	ID      ProfileID
	Name    string
	Surname string
}

// City город
type City struct {
	ID   uint64
	Name string
}

// Interest интерес
type Interest struct {
	ID   uint64
	Name string
}

func (p Profile) BirthDateFmt() string {
	return p.BirthDate.Format("2006-01-02")
}

func (p Profile) HasInterest(id uint64) bool {
	for _, interest := range p.Interests {
		if interest.ID == id {
			return true
		}
	}
	return false
}
