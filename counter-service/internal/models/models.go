package models

import "time"

type (
	ProfileID uint64

	UnreadMessagesCounterKey struct {
		From ProfileID `json:"from"`
		To   ProfileID `json:"to"`
	}

	Action uint8

	AlterCounterRequestStatus uint8

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

	Counter struct {
		Count int `json:"count"`
	}
)

const (
	ActionCreated Action = iota
	ActionRead
)

const (
	StatusOK AlterCounterRequestStatus = iota
	StatusFailed
)
