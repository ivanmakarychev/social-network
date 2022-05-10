package presentation

import (
	"log"
	"net/http"

	"github.com/ivanmakarychev/social-network/internal/authorization"
	"github.com/ivanmakarychev/social-network/internal/repository"
)

// App обрабатывает запросы пользователей
type App struct {
	authManager       authorization.Manager
	profileProvider   repository.ProfileRepo
	citiesProvider    repository.CitiesRepository
	interestsProvider repository.InterestsRepository
	friendsRepo       repository.FriendsRepo
}

func NewApp(
	authManager authorization.Manager,
	profileProvider repository.ProfileRepo,
	citiesProvider repository.CitiesRepository,
	interestsProvider repository.InterestsRepository,
	friendsRepo repository.FriendsRepo,
) *App {
	return &App{
		authManager:       authManager,
		profileProvider:   profileProvider,
		citiesProvider:    citiesProvider,
		interestsProvider: interestsProvider,
		friendsRepo:       friendsRepo,
	}
}

func (a *App) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", a.Register)
	mux.HandleFunc("/my/profile", a.BasicAuth(a.MyProfile))
	mux.HandleFunc("/profile", a.Profile)
	mux.HandleFunc("/make-friend", a.BasicAuth(a.MakeFriend))
	mux.HandleFunc("/confirm-friendship", a.BasicAuth(a.ConfirmFriendship))
	mux.HandleFunc("/", a.Home)

	srv := &http.Server{
		Addr:    ":80",
		Handler: mux,
	}

	log.Printf("start server on %s", srv.Addr)
	return srv.ListenAndServe()
}
