package model

import "time"

type Message struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	Recipient string    `json:"recipient"`
	IsSent    bool      `json:"isSent"`
	SentAt    time.Time `json:"sentAt,omitempty"`
	MessageID string    `json:"messageId,omitempty"` // Comes from Webhook Response
}

type ActionType string

const (
	ActionStart ActionType = "start"

	ActionStop ActionType = "stop"
)

func (a ActionType) IsValid() bool {
	switch a {
	case ActionStart, ActionStop:
		return true
	}
	return false
}

type StartStopRequest struct {
	Action ActionType `json:"action" validate:"oneof=start stop"`
}

type StartStopResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SentMessagesResponse struct {
	Messages []Message `json:"messages"`
	Count    int       `json:"count"`
}

type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"

	StatusStopped ServiceStatus = "stopped"
)
