package presentation

import (
	"log"
	"net/http"

	"github.com/ivanmakarychev/social-network/internal/repository"
)

// FindProfiles поиск профилей
func (a *App) FindProfiles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var findProfilesRq repository.FindProfilesRequest

	namePrefixes := query["first_name"]
	if len(namePrefixes) > 0 {
		findProfilesRq.NamePrefix = namePrefixes[0]
	}

	surnamePrefixes := query["surname"]
	if len(surnamePrefixes) > 0 {
		findProfilesRq.SurnamePrefix = surnamePrefixes[0]
	}

	profiles, err := a.profileProvider.FindProfiles(findProfilesRq)
	if err != nil {
		handleError("profile finder", "find profiles", err, w)
		return
	}

	data := ViewData{
		Title: "Результат поиска",
		Data: Profiles{
			Profiles: profiles,
		},
	}

	tmpl, err := loadTemplate("found_profiles.html")
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
