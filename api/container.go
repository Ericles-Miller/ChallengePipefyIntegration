package api

import (
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/repositories"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildClientController(pool *pgxpool.Pool) *clients.ClientController {
	repo    := repositories.NewClientRepository(pool)
	service := services.NewClientService(repo)
	return clients.NewClientController(service)
}
