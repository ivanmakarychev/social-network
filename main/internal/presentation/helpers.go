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
		"now":         time.Now,
		"printStatus": printStatus,
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

func (a *App) tryGetOwnerProfileFromContext(r *http.Request) (models.Profile, bool) {
	p, err := a.getOwnerProfileFromContext(r)
	return p, err == nil
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

func loadAndExecuteTemplate(template string, data any, w http.ResponseWriter) {
	tmpl, err := loadTemplate(template)
	if err != nil {
		log.Printf("failed to load template %q: %s\n", template, err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println("bad template ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func onlyMethod(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case method:
			h(w, r)
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
	}
}

func onlyPOST(h http.HandlerFunc) http.HandlerFunc {
	return onlyMethod(http.MethodPost, h)
}

func (a *App) Success(w http.ResponseWriter, _ *http.Request) {
	loadAndExecuteTemplate("success.html", nil, w)
}

func printStatus(status int) string {
	switch status {
	case 0:
		return "отправлено"
	case 1:
		return "доставлено"
	case 2:
		return "прочитано"
	default:
		return "неизвестный статус"
	}
}
