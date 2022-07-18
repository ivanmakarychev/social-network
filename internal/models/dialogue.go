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
		ID     MessageID
		Author ProfileID
		TS     time.Time
		Text   string
	}

	DialogueData struct {
		Dialogue
		Messages []*Message
	}

	MessageData struct {
		DialogueID
		*Message
	}
)
