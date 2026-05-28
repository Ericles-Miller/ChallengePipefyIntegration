package repositories

import (
	"context"
	clientdb "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/db"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientRepository interface {
	Create(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error)
	GetByEmail(ctx context.Context, email string) (*models.ClientResponse, error)
}

type clientRepository struct {
	queries *clientdb.Queries
}

func NewClientRepository(pool *pgxpool.Pool) ClientRepository {
	return &clientRepository{queries: clientdb.New(pool)}
}

func (r *clientRepository) Create(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error) {
	client, err := r.queries.CreateClient(ctx, clientdb.CreateClientParams{
		Name:           req.Name,
		Email:          req.Email,
		RequestType:    req.RequestType,
		PatrimonyValue: req.PatrimonyValue,
	})

	if err != nil {
		return nil, err
	}

	return models.ToResponse(client), nil
}

func (r *clientRepository) GetByEmail(ctx context.Context, email string) (*models.ClientResponse, error) {
	client, err := r.queries.GetClientByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return models.ToResponse(client), nil
}



