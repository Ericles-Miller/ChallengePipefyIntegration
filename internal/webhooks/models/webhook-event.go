package models

import (
	"time"

)

type WebhookEvent struct {
	EventID     string    `json:"event_id"`
	CardID      string    `json:"card_id"`
	ClientEmail string    `json:"client_email"`
	ProcessedAt time.Time `json:"processed_at"`
}

