package services

import (
	"context"
	"errors"

	clientModels "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/models"
	clientRepositories "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/repositories"
	webhookModels "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/models"
	webhookRepositories "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/repositories"
	AppError "github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/appError"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/pipefy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const highPriorityThreshold = 200_000.0

type txBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type WebhookService interface {
	ProcessEvent(ctx context.Context, req webhookModels.WebhookEventRequest) (*webhookModels.WebhookEventResponse, error)
}

type webhookService struct {
	webhookRepo webhookRepositories.WebhookRepository
	clientRepo  clientRepositories.ClientRepository
	pipefy      pipefy.PipefyClient
	db          txBeginner
}

func NewWebhookService(
	webhookRepo webhookRepositories.WebhookRepository,
	clientRepo clientRepositories.ClientRepository,
	pipefy pipefy.PipefyClient,
	pool *pgxpool.Pool,
) WebhookService {
	return &webhookService{webhookRepo: webhookRepo, clientRepo: clientRepo, pipefy: pipefy, db: pool}
}

func (s *webhookService) ProcessEvent(ctx context.Context, req webhookModels.WebhookEventRequest) (*webhookModels.WebhookEventResponse, error) {
	_, err := s.webhookRepo.GetEventByID(ctx, req.EventID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, AppError.New("failed to check existing event", AppError.ErrInternalServer)
	}
	if err == nil {
		return nil, AppError.New("event already processed", AppError.ErrBadRequest)
	}

	client, err := s.clientRepo.GetByEmail(ctx, req.ClientEmail)
	if err != nil {
		return nil, AppError.New("client not found", AppError.ErrNotFound)
	}

	priority := clientModels.PriorityNormal
	if client.PatrimonyValue >= highPriorityThreshold {
		priority = clientModels.PriorityHigh
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, AppError.New("failed to start transaction", AppError.ErrInternalServer)
	}
	defer tx.Rollback(ctx)

	updatedClient, err := s.clientRepo.WithTx(tx).UpdateClient(ctx, req.ClientEmail, clientModels.StatusProcessed, priority)
	if err != nil {
		return nil, AppError.New("failed to update client", AppError.ErrInternalServer)
	}

	if err := s.pipefy.MoveCardToProcessed(ctx, req.CardID); err != nil {
		if errors.Is(err, pipefy.ErrUnauthorized) {
			return nil, AppError.New("invalid Pipefy credentials", AppError.ErrUnauthorized)
		}
		return nil, AppError.New("failed to move Pipefy card to phase", AppError.ErrInternalServer)
	}

	if err := s.pipefy.UpdateCardField(ctx, req.CardID, pipefy.FieldPriority, string(priority)); err != nil {
		if errors.Is(err, pipefy.ErrUnauthorized) {
			return nil, AppError.New("invalid Pipefy credentials", AppError.ErrUnauthorized)
		}
		return nil, AppError.New("failed to update Pipefy card priority", AppError.ErrInternalServer)
	}

	event, err := s.webhookRepo.WithTx(tx).InsertEvent(ctx, req.EventID, req.CardID, req.ClientEmail)
	if err != nil {
		return nil, AppError.New("failed to save event", AppError.ErrInternalServer)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, AppError.New("failed to commit transaction", AppError.ErrInternalServer)
	}

	return &webhookModels.WebhookEventResponse{
		EventID:     event.EventID,
		ClientEmail: updatedClient.Email,
		Status:      updatedClient.Status,
		Priority:    updatedClient.Priority,
		ProcessedAt: event.ProcessedAt.Time,
	}, nil
}
