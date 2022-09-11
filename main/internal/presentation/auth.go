package presentation

import (
	"context"
	"errors"
	"net/http"

	"github.com/ivanmakarychev/social-network/internal/authorization"
	"github.com/ivanmakarychev/social-network/internal/models"
)

func (a *App) authorizeAndGetOwner(r *http.Request) (models.Profile, error) {
	username, password, ok := r.BasicAuth()
	if ok {
		userID, err := a.authManager.GetUserID(authorization.LoginData{
			Login:    username,
			Password: password,
		})
		if err != nil {
			return models.Profile{}, err
		}
		profile, err := a.profileProvider.GetProfile(userID)
		if err != nil {
			return models.Profile{}, err
		}
		return profile, nil
	}
	return models.Profile{}, errors.New("basic auth not provided")
}

func (a *App) BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		profile, err := a.authorizeAndGetOwner(r)
		if err == nil {
			r = r.WithContext(context.WithValue(r.Context(), "user", profile))
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Зарегистрируйтесь на странице /register", http.StatusUnauthorized)
	}
}

func (a *App) BasicAuthOptional(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		profile, err := a.authorizeAndGetOwner(r)
		if err == nil {
			r = r.WithContext(context.WithValue(r.Context(), "user", profile))
		}
		next.ServeHTTP(w, r)
	}
}
