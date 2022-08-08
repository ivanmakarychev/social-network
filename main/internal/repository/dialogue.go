package repository

import (
	"github.com/ivanmakarychev/social-network/internal/models"
)

type DialogueRepository interface {
	GetMessages(id models.DialogueID) ([]*models.Message, error)
	SaveMessage(msg *models.MessageData) error
}
