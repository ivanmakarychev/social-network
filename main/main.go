package main

import (
	"log"
	"os"

	"github.com/ivanmakarychev/social-network/internal/services"

	"github.com/ivanmakarychev/social-network/internal/tape"

	"github.com/ivanmakarychev/social-network/internal/authorization"
	"github.com/ivanmakarychev/social-network/internal/config"
	"github.com/ivanmakarychev/social-network/internal/presentation"
	"github.com/ivanmakarychev/social-network/internal/repository"
)

func main() {
	cfg, err := config.ReadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatal("failed to read config: ", err)
	}

	db := repository.NewMySQLCluster(cfg.Database, nil)
	err = db.Init()
	if err != nil {
		log.Fatal("failed to create MySQL cluster: ", err)
	}
	defer db.Close()

	citiesRepo := repository.NewCitiesRepositoryImpl(db)
	interestRepo := repository.NewInterestsRepositoryImpl(db)
	friendsRepo := repository.NewFriendsRepoImpl(db)
	profileRepo := repository.NewProfileRepoImpl(db, friendsRepo)
	authManager := authorization.NewManagerImpl(db)
	updatesRepo := repository.NewClusterUpdatesRepo(db)

	updatesQueue := tape.NewBroadcastQueue(cfg.Updates.QueueConnStr)
	err = updatesQueue.Init()
	if err != nil {
		log.Fatal("failed to init updates queue: ", err)
	}
	defer updatesQueue.Close()

	updatesDirectQueue := tape.NewDirectQueue(cfg.Updates.QueueConnStr)
	err = updatesDirectQueue.Init()
	if err != nil {
		log.Fatal("failed to init updates direct queue: ", err)
	}
	defer updatesDirectQueue.Close()

	tapeProvider := tape.NewCachingProvider(
		cfg.Updates,
		updatesRepo,
		updatesQueue,
		tape.NewUpdatesSubscriptionManagerImpl(updatesDirectQueue),
	)
	subscription := tape.NewSubscriptionImpl(updatesRepo)

	dialogueService := services.NewDialogueService(cfg.DialogueService.Connection)

	app := presentation.NewApp(
		cfg.Server,
		authManager,
		profileRepo,
		citiesRepo,
		interestRepo,
		friendsRepo,
		dialogueService,
		tapeProvider,
		subscription,
		tape.NewRouterPublisher(
			updatesRepo,
			tape.NewCompositeUpdatesRouter(
				updatesQueue,
				updatesDirectQueue,
			),
			cfg.Updates,
		),
	)

	log.Fatal(app.Run())
}
