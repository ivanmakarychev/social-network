package saga

import (
	"context"
	"log"

	"github.com/ivanmakarychev/social-network/counter-service/internal/models"
	"github.com/ivanmakarychev/social-network/counter-service/internal/repository"
)

type (
	Saga struct {
		In          <-chan *models.AlterCounterRequest
		Out         chan<- *models.AlterCounterResult
		CounterRepo repository.CounterRepository
	}
)

func (s *Saga) Run(ctx context.Context) {
	go func() {
		for request := range s.In {
			sign := 1
			if request.Action == models.ActionRead {
				sign = -1
			}
			result := &models.AlterCounterResult{
				Key:               request.Key,
				Action:            request.Action,
				MessageTimestamps: request.MessageTimestamps,
				Status:            models.StatusOK,
			}
			err := s.CounterRepo.Increment(ctx, request.Key, sign*len(request.MessageTimestamps))
			if err != nil {
				result.Status = models.StatusFailed
				log.Println("failed to increment counter:", err.Error())
			}
			s.Out <- result
		}
	}()
}
