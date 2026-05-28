package api

import (
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients"
	clientRepositories "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/repositories"
	clientServices "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/services"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks"
	webhookRepositories "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/repositories"
	webhookServices "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildClientController(pool *pgxpool.Pool) *clients.ClientController {
	repo    := clientRepositories.NewClientRepository(pool)
	service := clientServices.NewClientService(repo)
	return clients.NewClientController(service)
}

func buildWebhookController(pool *pgxpool.Pool) *webhooks.WebhookController {
	clientRepo  := clientRepositories.NewClientRepository(pool)
	webhookRepo := webhookRepositories.NewWebhookRepository(pool)
	service     := webhookServices.NewWebhookService(webhookRepo, clientRepo)
	return webhooks.NewWebhookController(service)
}
