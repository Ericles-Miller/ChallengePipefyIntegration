package api

import (
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients"
	clientRepositories "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/repositories"
	clientServices "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/services"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks"
	webhookRepositories "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/repositories"
	webhookServices "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/services"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/pipefy"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildClientController(pool *pgxpool.Pool) *clients.ClientController {
	repo        := clientRepositories.NewClientRepository(pool)
	pipefyClient := pipefy.NewPipefyClient()
	service     := clientServices.NewClientService(repo, pipefyClient)
	return clients.NewClientController(service)
}

func buildWebhookController(pool *pgxpool.Pool) *webhooks.WebhookController {
	clientRepo   := clientRepositories.NewClientRepository(pool)
	webhookRepo  := webhookRepositories.NewWebhookRepository(pool)
	pipefyClient := pipefy.NewPipefyClient()
	service      := webhookServices.NewWebhookService(webhookRepo, clientRepo, pipefyClient)
	return webhooks.NewWebhookController(service)
}
