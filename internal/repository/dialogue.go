package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/ivanmakarychev/social-network/internal/models"
)

type DialogueRepository interface {
	GetMessages(id models.DialogueID) ([]*models.Message, error)
	SaveMessage(msg *models.MessageData) error
}

type PostgreDialogueRepository struct {
	db DialogueDB
}

func NewPostgreDialogueRepository(db DialogueDB) *PostgreDialogueRepository {
	return &PostgreDialogueRepository{
		db: db,
	}
}

func (p *PostgreDialogueRepository) GetMessages(id models.DialogueID) ([]*models.Message, error) {
	id = normalizeDialogueID(id)
	query, args, err := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Select("message_id", "profile_id_author", "ts", "text").
		From("dialogue_message").
		Where(squirrel.Eq{
			"profile_id_1": id.ProfileID1,
			"profile_id_2": id.ProfileID2,
		}).OrderBy("message_id").ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.db.GetConn().Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]*models.Message, 0)
	for rows.Next() {
		m := models.Message{}
		err = rows.Scan(&m.ID, &m.Author, &m.TS, &m.Text)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &m)
	}
	return messages, nil
}

func (p *PostgreDialogueRepository) SaveMessage(msg *models.MessageData) error {
	dialogueID := normalizeDialogueID(msg.DialogueID)
	query, args, err := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Insert("dialogue_message").
		Columns("profile_id_1", "profile_id_2", "profile_id_author", "text").
		Values(dialogueID.ProfileID1, dialogueID.ProfileID2, msg.Author, msg.Text).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.db.GetConn().Exec(context.Background(), query, args...)
	return err
}

func normalizeDialogueID(id models.DialogueID) models.DialogueID {
	if id.ProfileID1 < id.ProfileID2 {
		return id
	}
	return models.DialogueID{
		ProfileID1: id.ProfileID2,
		ProfileID2: id.ProfileID1,
	}
}
