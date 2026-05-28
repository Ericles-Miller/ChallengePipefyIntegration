package models

import "time"

type WebhookEventRequest struct {
	EventID     string    `json:"event_id"      binding:"required,max=255"`
	CardID      string    `json:"card_id"       binding:"required,max=255"`
	ClientEmail string    `json:"cliente_email" binding:"required,email,max=255"`
	Timestamp   time.Time `json:"timestamp"     binding:"required"`
}
