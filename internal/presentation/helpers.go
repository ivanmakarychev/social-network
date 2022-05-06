package presentation

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/ivanmakarychev/social-network/internal/models"
)

const (
	templatePath = "internal/presentation/templates/"
	baseTemplate = templatePath + "base.html"

	errorMessage = "Упс... что-то пошло не так. Напишите нам, а то мы об этом не узнаем"
)

var (
	funcMap = template.FuncMap{
		"now": time.Now,
	}
)

func loadTemplate(filename string) (*template.Template, error) {
	return template.New("").
		Funcs(funcMap).
		ParseFiles(
			templatePath+filename,
			baseTemplate,
		)
}

func (a *App) getOwnerProfileFromContext(r *http.Request) (models.Profile, error) {
	v := r.Context().Value("user")
	if v == nil {
		return a.authorizeAndGetOwner(r)
	}
	p, ok := v.(models.Profile)
	if ok {
		return p, nil
	}
	return models.Profile{}, errors.New("key 'user' in context has wrong type")
}

func (a *App) getUserProfile(id models.ProfileID) (models.Profile, error) {
	return a.profileProvider.GetProfile(id)
}

func handleError(actor, action string, err error, w http.ResponseWriter) {
	log.Println(
		fmt.Sprintf("[error] %s failed to %s: %s",
			actor,
			action,
			err,
		))
	http.Error(w, errorMessage, http.StatusInternalServerError)
}
