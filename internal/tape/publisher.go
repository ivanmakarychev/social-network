package tape

import "github.com/ivanmakarychev/social-network/internal/models"

type (
	Publisher interface {
		Publish(u *models.Update) error
	}

	QueuePublisher struct {
	}
)

func NewQueuePublisher() *QueuePublisher {
	return &QueuePublisher{}
}

func (q *QueuePublisher) Publish(u *models.Update) error {
	//todo impl
	return nil
}
