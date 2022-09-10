package main

import (
	"context"
	"log"
	"os"

	"github.com/ivanmakarychev/social-network/counter-service/internal/api"
	"github.com/ivanmakarychev/social-network/counter-service/internal/config"
	"github.com/ivanmakarychev/social-network/counter-service/internal/repository"
	"github.com/ivanmakarychev/social-network/counter-service/internal/saga"
)

func main() {
	cfg, err := config.ReadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatal("failed to read config: ", err)
	}

	counterRepo, err := repository.MakeRedisCounterRepository(cfg.Redis)
	if err != nil {
		log.Fatal("failed to make counter repo: ", err)
	}

	s := &saga.Saga{
		CounterRepo: counterRepo,
	}

	consumer := saga.NewConsumerImpl(cfg.MQ.ConnStr)
	s.In, err = consumer.Init()
	if err != nil {
		log.Fatal("failed to init consumer: ", err)
	}

	publisher := saga.NewPublisherImpl(cfg.MQ.ConnStr)
	s.Out, err = publisher.Init()
	if err != nil {
		log.Fatal("failed to init publisher: ", err)
	}

	s.Run(context.Background())

	app := api.NewAPI(
		cfg.Service,
		counterRepo,
	)

	log.Fatal(app.Run())
}
