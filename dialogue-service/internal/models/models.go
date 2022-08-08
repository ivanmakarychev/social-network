package models

import "time"

type (
	ProfileID uint64

	DialogueID struct {
		ProfileID1 ProfileID `json:"profile_id_1"`
		ProfileID2 ProfileID `json:"profile_id_2"`
	}

	MessageID uint64

	Message struct {
		ID     MessageID `json:"id"`
		Author ProfileID `json:"author"`
		TS     time.Time `json:"ts"`
		Text   string    `json:"text"`
	}

	DialogueData struct {
		DialogueID `json:"dialogue_id"`
		Messages   []*Message `json:"messages"`
	}

	MessageData struct {
		DialogueID
		*Message
	}

	PostMessageRq struct {
		From ProfileID `json:"from"`
		To   ProfileID `json:"to"`
		Text string    `json:"text"`
	}
)
