package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/models"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/repositories"
	AppError "github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/appError"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/pipefy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type ClientService interface {
	CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error)
	GetByEmail(ctx context.Context, email string) (*models.ClientResponse, error)
}

type clientService struct {
	repo   repositories.ClientRepository
	pipefy pipefy.PipefyClient
	db     txBeginner
}

func NewClientService(repo repositories.ClientRepository, pipefy pipefy.PipefyClient, pool *pgxpool.Pool) ClientService {
	return &clientService{repo: repo, pipefy: pipefy, db: pool}
}

func (s *clientService) CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error) {
	_, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, AppError.New("failed to check existing client", AppError.ErrInternalServer)
	}
	if err == nil {
		return nil, AppError.New(fmt.Sprintf("client with email '%s' already exists", req.Email), AppError.ErrBadRequest)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, AppError.New("failed to start transaction", AppError.ErrInternalServer)
	}
	defer tx.Rollback(ctx)

	client, err := s.repo.WithTx(tx).Create(ctx, req)
	if err != nil {
		return nil, AppError.New("failed to create client", AppError.ErrInternalServer)
	}

	if _, err := s.pipefy.CreateCard(ctx, req.Name, req.Email, req.RequestType, req.PatrimonyValue); err != nil {
		if errors.Is(err, pipefy.ErrUnauthorized) {
			return nil, AppError.New("invalid Pipefy credentials", AppError.ErrUnauthorized)
		}
		return nil, AppError.New("failed to create Pipefy card", AppError.ErrInternalServer)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, AppError.New("failed to commit transaction", AppError.ErrInternalServer)
	}

	return client, nil
}

func (s *clientService) GetByEmail(ctx context.Context, email string) (*models.ClientResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, AppError.New(fmt.Sprintf("client with email '%s' not found", email), AppError.ErrNotFound)
	}

	return user, nil
}
