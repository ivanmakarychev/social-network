package models

import "time"

type (
	DialogueID struct {
		ProfileID1 ProfileID
		ProfileID2 ProfileID
	}

	Dialogue struct {
		ID   DialogueID
		Who  ProfileID
		With ProfileMain
	}

	MessageID uint64

	Message struct {
		ID     MessageID `json:"id"`
		Author ProfileID `json:"author"`
		TS     time.Time `json:"ts"`
		Text   string    `json:"text"`
	}

	DialogueData struct {
		Dialogue
		Messages []*Message
	}

	MessageData struct {
		To ProfileID
		*Message
	}
)
