package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/saga"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/api"
	"github.com/ivanmakarychev/social-network/dialogue-service/internal/config"
	"github.com/ivanmakarychev/social-network/dialogue-service/internal/repository"
)

func main() {
	cfg, err := config.ReadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatal("failed to read config: ", err)
	}

	go func() {
		promHandler := http.NewServeMux()
		promHandler.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":2112", promHandler)
		if err != nil {
			log.Fatal("failed to start metrics handler on port 2112:", err)
		}
	}()

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

	s := &saga.Saga{
		DialogueRepository: dialogueRepo,
	}

	consumer := saga.NewConsumerImpl(cfg.MQ.ConnStr)
	s.In, err = consumer.Init()
	if err != nil {
		log.Fatal("failed to init consumer: ", err)
	}
	defer consumer.Close()

	s.Publisher = saga.NewPublisherImpl(cfg.MQ.ConnStr)
	err = s.Publisher.Init()
	if err != nil {
		log.Fatal("failed to init publisher: ", err)
	}
	defer s.Publisher.Close()

	s.Run()

	app := api.NewAPI(
		cfg.Server,
		dialogueRepo,
		s,
		metrics.New(),
	)

	log.Fatal(app.Run())
}
