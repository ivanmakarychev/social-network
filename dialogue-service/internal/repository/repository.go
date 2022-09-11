package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/ivanmakarychev/social-network/dialogue-service/internal/models"
)

type DialogueRepository interface {
	GetMessages(id models.DialogueID) ([]*models.Message, error)
	SaveMessage(msg *models.MessageData) error
	UpdateMessageStatus(key models.MessageKey, status models.MessageStatus) error
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
		Select("profile_id_author", "ts", "text", "status").
		From("dialogue_message").
		Where(squirrel.And{
			squirrel.Eq{
				"profile_id_1": id.ProfileID1,
				"profile_id_2": id.ProfileID2,
			},
			squirrel.NotEq{
				"status": -1,
			},
		}).OrderBy("ts").ToSql()
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
		err = rows.Scan(&m.Author, &m.TS, &m.Text, &m.Status)
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
		Columns("profile_id_1", "profile_id_2", "profile_id_author", "text", "ts", "status").
		Values(dialogueID.ProfileID1, dialogueID.ProfileID2, msg.Author, msg.Text, msg.TS, msg.Status).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.db.GetConn().Exec(context.Background(), query, args...)
	return err
}

func (p *PostgreDialogueRepository) UpdateMessageStatus(key models.MessageKey, status models.MessageStatus) error {
	dialogueID := normalizeDialogueID(models.DialogueID{
		ProfileID1: key.From,
		ProfileID2: key.To,
	})
	query, args, err := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Update("dialogue_message").
		Set("status", status).
		Where(squirrel.Eq{
			"profile_id_1":      dialogueID.ProfileID1,
			"profile_id_2":      dialogueID.ProfileID2,
			"ts":                key.TS,
			"profile_id_author": key.From,
		}).
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
