package presentation

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ivanmakarychev/social-network/internal/models"
)

// ProfileEditData данные для редактирования профиля
type ProfileEditData struct {
	Saved     bool
	Profile   models.Profile
	Cities    []models.City
	Interests []models.Interest
}

type OtherProfileData struct {
	Profile            models.Profile
	FriendshipProposed bool
}

// MyProfile домашняя страница пользователя
func (a *App) MyProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getMyProfile(w, r)
	case http.MethodPost:
		a.saveMyProfile(w, r)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

// Profile страница другого пользователя
func (a *App) Profile(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ids, present := query["id"]
	if !present || len(ids) != 1 {
		a.NotFound(w, r)
		return
	}
	idStr := ids[0]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		a.NotFound(w, r)
		return
	}

	profile, err := a.getUserProfile(models.ProfileID(id))
	if err != nil {
		log.Println("failed to find profile ", id, ": ", err)
		a.NotFound(w, r)
		return
	}

	_, friendshipApplicationCreated := query["friendship_proposed"]

	data := ViewData{
		Title: fmt.Sprintf("%s %s", profile.Name, profile.Surname),
		Data: OtherProfileData{
			Profile:            profile,
			FriendshipProposed: friendshipApplicationCreated,
		},
	}
	tmpl, err := loadTemplate("profile.html")
	if err != nil {
		log.Println("failed to load template: ", err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println("bad template ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) getMyProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := a.getOwnerProfileFromContext(r)
	if err != nil {
		log.Println("failed to get user profile: ", err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	cities, err := a.citiesProvider.GetCities()
	if err != nil {
		log.Println("failed to get cities: ", err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	interests, err := a.interestsProvider.GetInterests()
	if err != nil {
		log.Println("failed to get interests: ", err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	data := ViewData{
		Title: fmt.Sprintf("%s %s - это ты", profile.Name, profile.Surname),
		Data: ProfileEditData{
			Profile:   profile,
			Cities:    cities,
			Interests: interests,
		},
	}
	tmpl, err := loadTemplate("my_profile.html")
	if err != nil {
		log.Println("failed to load template: ", err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println("bad template ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) saveMyProfile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		handleError("profile", "parse form", err, w)
		return
	}

	profile, err := a.getOwnerProfileFromContext(r)
	if err != nil {
		log.Println("failed to get user profile: ", err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	form := r.PostForm

	if len(form["id"]) != 1 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	profileID, err := strconv.ParseUint(form["id"][0], 10, 64)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if models.ProfileID(profileID) != profile.ID {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	if len(form["name"]) != 1 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	profile.Name = form["name"][0]
	if len(profile.Name) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if len(profile.Name) > 32 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if len(form["surname"]) != 1 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	profile.Surname = form["surname"][0]
	if len(profile.Name) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if len(profile.Surname) > 32 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if len(form["birth_date"]) != 1 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	profile.BirthDate, err = time.Parse("2006-01-02", form["birth_date"][0])
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if len(form["city"]) != 1 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	cityID, err := strconv.ParseUint(form["city"][0], 10, 64)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	cities, err := a.citiesProvider.GetCities()
	if err != nil {
		handleError("profile", "get cities", err, w)
		return
	}
	cityFound := false
	for _, city := range cities {
		if city.ID == cityID {
			cityFound = true
			profile.City = city
			break
		}
	}
	if !cityFound {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	interestsStr := form["interests"]
	if len(interestsStr) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	interestIDs := make([]uint64, 0, len(interestsStr))
	for _, interestStr := range interestsStr {
		id, err := strconv.ParseUint(interestStr, 10, 64)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		interestIDs = append(interestIDs, id)
	}
	interests, err := a.interestsProvider.GetInterests()
	if err != nil {
		handleError("profile", "get interests", err, w)
		return
	}
	profile.Interests = make([]models.Interest, 0, len(interestIDs))
Loop:
	for _, interest := range interests {
		for _, id := range interestIDs {
			if interest.ID == id {
				profile.Interests = append(profile.Interests, interest)
				if len(profile.Interests) == len(interestIDs) {
					break Loop
				}
				continue Loop
			}
		}
	}
	if len(profile.Interests) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	err = a.profileProvider.SaveProfile(profile)
	if err != nil {
		handleError("profile", "save profile", err, w)
		return
	}

	data := ViewData{
		Title: fmt.Sprintf("%s %s - это ты", profile.Name, profile.Surname),
		Data: ProfileEditData{
			Saved:     true,
			Profile:   profile,
			Cities:    cities,
			Interests: interests,
		},
	}
	tmpl, err := loadTemplate("my_profile.html")
	if err != nil {
		handleError("profile", "load template", err, w)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleError("profile", "execute template", err, w)
		return
	}
}
