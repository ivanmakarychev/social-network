package presentation

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ivanmakarychev/social-network/internal/authorization"
	"github.com/ivanmakarychev/social-network/internal/config"
	"github.com/ivanmakarychev/social-network/internal/repository"
)

// App обрабатывает запросы пользователей
type App struct {
	cfg               config.Server
	authManager       authorization.Manager
	profileProvider   repository.ProfileRepo
	citiesProvider    repository.CitiesRepository
	interestsProvider repository.InterestsRepository
	friendsRepo       repository.FriendsRepo
	dialogueRepo      repository.DialogueRepository
}

func NewApp(
	cfg config.Server,
	authManager authorization.Manager,
	profileProvider repository.ProfileRepo,
	citiesProvider repository.CitiesRepository,
	interestsProvider repository.InterestsRepository,
	friendsRepo repository.FriendsRepo,
	dialogueRepo repository.DialogueRepository,
) *App {
	return &App{
		cfg:               cfg,
		authManager:       authManager,
		profileProvider:   profileProvider,
		citiesProvider:    citiesProvider,
		interestsProvider: interestsProvider,
		friendsRepo:       friendsRepo,
		dialogueRepo:      dialogueRepo,
	}
}

func (a *App) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", a.Register)
	mux.HandleFunc("/my/profile", a.BasicAuth(a.MyProfile))
	mux.HandleFunc("/profile", a.Profile)
	mux.HandleFunc("/profiles", a.FindProfiles)
	mux.HandleFunc("/make-friend", a.BasicAuth(a.MakeFriend))
	mux.HandleFunc("/confirm-friendship", a.BasicAuth(a.ConfirmFriendship))

	mux.HandleFunc("/dialogue", a.BasicAuth(a.ShowDialogue))
	mux.HandleFunc("/dialogue/message/send", a.BasicAuth(a.SendMessage))

	mux.HandleFunc("/", a.Home)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", a.cfg.Port),
		Handler: mux,
	}

	log.Printf("start server on %s", srv.Addr)
	return srv.ListenAndServe()
}
