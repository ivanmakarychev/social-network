package main

import (
	"context"
	"log"
	"os"

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

	updatesQueue := tape.NewQueueImpl(cfg.Updates.QueueConnStr)
	err = updatesQueue.Init()
	if err != nil {
		log.Fatal("failed to init updates queue: ", err)
	}
	defer updatesQueue.Close()

	tapeProvider := tape.NewCachingProvider(cfg.Updates, updatesRepo, updatesQueue)
	subscription := tape.NewSubscriptionImpl(updatesRepo)

	dialogueDB, err := repository.NewShardedDialogueDB(cfg.DialogueDatabase, nil)
	if err != nil {
		log.Fatal("failed to create dialogue DB: ", err)
	}
	err = dialogueDB.Init(context.Background())
	if err != nil {
		log.Fatal("failed to init dialogue DB: ", err)
	}
	defer dialogueDB.Close()
	dialogueRepo := repository.NewPostgreDialogueRepository(dialogueDB)

	app := presentation.NewApp(
		cfg.Server,
		authManager,
		profileRepo,
		citiesRepo,
		interestRepo,
		friendsRepo,
		dialogueRepo,
		tapeProvider,
		subscription,
		tape.NewQueuePublisher(updatesRepo, updatesQueue, cfg.Updates),
	)

	log.Fatal(app.Run())
}
