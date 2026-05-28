package models

import (
	"time"
	clientModels "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/models"
)

type WebhookEventResponse struct {
	EventID     string                      `json:"event_id"`
	ClientEmail string                      `json:"cliente_email"`
	Status      clientModels.ClientStatus   `json:"status"`
	Priority    clientModels.ClientPriority `json:"prioridade"`
	ProcessedAt time.Time                   `json:"processed_at"`
}
