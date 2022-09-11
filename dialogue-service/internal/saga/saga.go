package saga

import (
	"log"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/models"
	"github.com/ivanmakarychev/social-network/dialogue-service/internal/repository"
)

type (
	Saga struct {
		In                 <-chan *models.AlterCounterResult
		Publisher          *PublisherImpl
		DialogueRepository repository.DialogueRepository
	}
)

func (s *Saga) UpdateMessages(rq *models.AlterCounterRequest) error {
	return s.Publisher.Publish(rq)
}

func (s *Saga) Run() {
	go func() {
		for alterCounterResult := range s.In {
			messageStatus := deriveMessageStatus(alterCounterResult.Action, alterCounterResult.Status)
			for _, ts := range alterCounterResult.MessageTimestamps {
				err := withRetry(
					func() error {
						return s.DialogueRepository.UpdateMessageStatus(
							models.MessageKey{
								From: alterCounterResult.Key.From,
								To:   alterCounterResult.Key.To,
								TS:   ts,
							},
							messageStatus,
						)
					},
					3,
				)
				if err != nil {
					log.Println("failed to update message status", err)
				}
			}
		}
	}()
}

func deriveMessageStatus(action models.Action, status models.AlterCounterRequestStatus) models.MessageStatus {
	switch action {
	case models.ActionCreated:
		switch status {
		case models.StatusOK:
			return models.MessageStatusDelivered
		case models.StatusFailed:
			return models.MessageStatusFailed
		}
	case models.ActionRead:
		switch status {
		case models.StatusOK:
			return models.MessageStatusRead
		case models.StatusFailed:
			return models.MessageStatusDelivered
		}
	}
	return models.MessageStatusCreated
}

func withRetry(f func() error, maxTimes int) error {
	var err error
	for i := 0; i < maxTimes; i++ {
		err = f()
		if err == nil {
			return nil
		}
	}
	return err
}
