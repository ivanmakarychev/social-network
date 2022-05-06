package presentation

import (
	"log"
	"net/http"

	"github.com/ivanmakarychev/social-network/internal/authorization"
	"github.com/ivanmakarychev/social-network/internal/models"
)

// Register страница регистрации
func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.registerGet(w, r)
		return
	case http.MethodPost:
		a.registerPost(w, r)
		return
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) registerGet(w http.ResponseWriter, _ *http.Request) {
	data := ViewData{
		Title: "Регистрация",
	}
	tmpl, err := loadTemplate("register.html")
	if err != nil {
		log.Println("bad template ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println("bad template ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) registerPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("failed to parse register form: ", err)
		http.Error(w, "something's got wrong", http.StatusBadRequest)
		return
	}

	logins := r.PostForm["login"]
	if len(logins) != 1 {
		http.Error(w, "provide login", http.StatusBadRequest)
		return
	}
	login := logins[0]
	if len(login) == 0 {
		http.Error(w, "provide login", http.StatusBadRequest)
		return
	}
	if len(login) > 32 {
		http.Error(w, "login's too long", http.StatusBadRequest)
		return
	}

	passwords := r.PostForm["password"]
	if len(passwords) != 1 {
		http.Error(w, "provide password", http.StatusBadRequest)
		return
	}
	password := passwords[0]
	if len(password) == 0 {
		http.Error(w, "provide password", http.StatusBadRequest)
		return
	}
	if len(password) > 32 {
		http.Error(w, "password's too long", http.StatusBadRequest)
		return
	}

	passwordConfirms := r.PostForm["password_confirm"]
	if len(passwordConfirms) != 1 {
		http.Error(w, "provide password confirm", http.StatusBadRequest)
		return
	}
	passwordConfirm := passwordConfirms[0]
	if len(passwordConfirms) == 0 {
		http.Error(w, "provide password confirm", http.StatusBadRequest)
		return
	}
	if passwordConfirm != password {
		http.Error(w, "password is not confirmed", http.StatusBadRequest)
		return
	}

	err = authorization.ValidatePassword(password)
	if err != nil {
		http.Error(w, "password is not proper", http.StatusBadRequest)
		return
	}

	var profileID models.ProfileID
	profileID, err = a.profileProvider.CreateProfileID()
	if err != nil {
		handleError("profile", "create profile id", err, w)
		return
	}

	err = a.authManager.SaveLogin(profileID, authorization.LoginData{
		Login:    login,
		Password: password,
	})
	if err != nil {
		log.Println("failed to save login and password: ", err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/my/profile", http.StatusFound)
}
