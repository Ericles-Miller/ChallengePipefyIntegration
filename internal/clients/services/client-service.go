package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/models"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/repositories"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/appError"
	"github.com/jackc/pgx/v5"
)

type ClientService interface {
	CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error)
	GetByEmail(ctx context.Context, email string) (*models.ClientResponse, error)
}

type clientService struct {
	repo repositories.ClientRepository
}

func NewClientService(repo repositories.ClientRepository) ClientService {
	return &clientService{repo: repo}
}

func (s *clientService) CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error) {
	_, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing client: %w", AppError.ErrInternalServer)
	}
	
	if err == nil {
		return nil, fmt.Errorf("client with email '%s' already exists: %w", req.Email, AppError.ErrBadRequest)
	}

	return s.repo.Create(ctx, req)
}

func (s *clientService) GetByEmail(ctx context.Context, email string) (*models.ClientResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get client with email '%s': %w", email, AppError.ErrNotFound)
	}

	return user, nil
}

