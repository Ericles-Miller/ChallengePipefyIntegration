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
)

type ClientService interface {
	CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error)
	GetByEmail(ctx context.Context, email string) (*models.ClientResponse, error)
}

type clientService struct {
	repo   repositories.ClientRepository
	pipefy pipefy.PipefyClient
}

func NewClientService(repo repositories.ClientRepository, pipefy pipefy.PipefyClient) ClientService {
	return &clientService{repo: repo, pipefy: pipefy}
}

func (s *clientService) CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error) {
	_, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, AppError.New("failed to check existing client", AppError.ErrInternalServer)
	}

	if err == nil {
		return nil, AppError.New(fmt.Sprintf("client with email '%s' already exists", req.Email), AppError.ErrBadRequest)
	}

	client, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, AppError.New("failed to create client", AppError.ErrInternalServer)
	}

	if _, err := s.pipefy.CreateCard(ctx, req.Name, req.Email, req.RequestType, req.PatrimonyValue); err != nil {
		return nil, AppError.New("failed to create Pipefy card", AppError.ErrInternalServer)
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

