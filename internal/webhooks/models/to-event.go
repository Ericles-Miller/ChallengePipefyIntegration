package models

import "time"

func ToEvent(req WebhookEventRequest) WebhookEvent {
	return WebhookEvent{
		EventID:     req.EventID,
		CardID:      req.CardID,
		ClientEmail: req.ClientEmail,
		ProcessedAt: time.Now(),
	}
}
