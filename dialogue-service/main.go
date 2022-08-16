package main

import (
	"context"
	"log"
	"os"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/api"
	"github.com/ivanmakarychev/social-network/dialogue-service/internal/config"
	"github.com/ivanmakarychev/social-network/dialogue-service/internal/repository"
)

func main() {
	cfg, err := config.ReadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatal("failed to read config: ", err)
	}

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

	app := api.NewAPI(
		cfg.Server,
		dialogueRepo,
	)

	log.Fatal(app.Run())
}
