package models

import "time"

type (
	ProfileID uint64

	DialogueID struct {
		ProfileID1 ProfileID `json:"profile_id_1"`
		ProfileID2 ProfileID `json:"profile_id_2"`
	}

	Message struct {
		Author ProfileID     `json:"author"`
		TS     time.Time     `json:"ts"`
		Text   string        `json:"text"`
		Status MessageStatus `json:"status"`
	}

	MessageStatus int

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

	MessageKey struct {
		From ProfileID `json:"from"`
		To   ProfileID `json:"to"`
		TS   time.Time `json:"ts"`
	}

	Action uint8

	AlterCounterRequestStatus uint8

	UnreadMessagesCounterKey struct {
		From ProfileID `json:"from"`
		To   ProfileID `json:"to"`
	}

	AlterCounterRequest struct {
		Key               UnreadMessagesCounterKey `json:"key"`
		Action            Action                   `json:"action"`
		MessageTimestamps []time.Time              `json:"message_timestamps"`
	}

	AlterCounterResult struct {
		Key               UnreadMessagesCounterKey  `json:"key"`
		Action            Action                    `json:"action"`
		MessageTimestamps []time.Time               `json:"message_timestamps"`
		Status            AlterCounterRequestStatus `json:"status"`
	}
)

const (
	MessageStatusCreated   MessageStatus = 0
	MessageStatusFailed    MessageStatus = -1
	MessageStatusDelivered MessageStatus = 1
	MessageStatusRead      MessageStatus = 2
)

const (
	ActionCreated Action = iota
	ActionRead
)

const (
	StatusOK AlterCounterRequestStatus = iota
	StatusFailed
)
