package services

import (
	"context"
	"errors"
	"testing"
	"time"

	clientModels "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/models"
	webhookdb "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/db"
	webhookModels "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/models"
	AppError "github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/appError"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/pipefy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- mock implementations ---

type mockWebhookClientRepo struct {
	getByEmailFn   func(ctx context.Context, email string) (*clientModels.ClientResponse, error)
	updateClientFn func(ctx context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error)
}

func (m *mockWebhookClientRepo) Create(ctx context.Context, req clientModels.CreateClientRequest) (*clientModels.ClientResponse, error) {
	panic("Create not expected in webhook service tests")
}

func (m *mockWebhookClientRepo) GetByEmail(ctx context.Context, email string) (*clientModels.ClientResponse, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockWebhookClientRepo) UpdateClient(ctx context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error) {
	return m.updateClientFn(ctx, email, status, priority)
}

type mockWebhookRepo struct {
	insertEventFn  func(ctx context.Context, eventID, cardID, clientEmail string) (webhookdb.WebhookEvent, error)
	getEventByIDFn func(ctx context.Context, eventID string) (webhookdb.WebhookEvent, error)
}

func (m *mockWebhookRepo) InsertEvent(ctx context.Context, eventID, cardID, clientEmail string) (webhookdb.WebhookEvent, error) {
	return m.insertEventFn(ctx, eventID, cardID, clientEmail)
}

func (m *mockWebhookRepo) GetEventByID(ctx context.Context, eventID string) (webhookdb.WebhookEvent, error) {
	return m.getEventByIDFn(ctx, eventID)
}

type mockWebhookPipefy struct {
	createCardFn          func(ctx context.Context, name, email, requestType string, patrimony float64) (*pipefy.CardResult, error)
	moveCardToProcessedFn func(ctx context.Context, cardID string) error
	updateCardFieldFn     func(ctx context.Context, cardID, fieldID, value string) error
}

func (m *mockWebhookPipefy) CreateCard(ctx context.Context, name, email, requestType string, patrimony float64) (*pipefy.CardResult, error) {
	return m.createCardFn(ctx, name, email, requestType, patrimony)
}

func (m *mockWebhookPipefy) MoveCardToProcessed(ctx context.Context, cardID string) error {
	return m.moveCardToProcessedFn(ctx, cardID)
}

func (m *mockWebhookPipefy) UpdateCardField(ctx context.Context, cardID, fieldID, value string) error {
	return m.updateCardFieldFn(ctx, cardID, fieldID, value)
}

// newSuccessfulPipefy returns a Pipefy mock that succeeds all calls.
func newSuccessfulPipefy() *mockWebhookPipefy {
	return &mockWebhookPipefy{
		moveCardToProcessedFn: func(_ context.Context, _ string) error { return nil },
		updateCardFieldFn:     func(_ context.Context, _, _, _ string) error { return nil },
	}
}

// --- tests ---

// TestProcessEvent_HighPriority verifies that a client with valor_patrimonio >= 200.000
// receives prioridade_alta after webhook processing.
func TestProcessEvent_HighPriority(t *testing.T) {
	req := webhookModels.WebhookEventRequest{
		EventID:     "evt_high",
		CardID:      "card_456",
		ClientEmail: "joao.silva@example.com",
		Timestamp:   time.Now(),
	}

	clientWithHighPatrimony := &clientModels.ClientResponse{
		ID:             "uuid-1",
		Email:          req.ClientEmail,
		PatrimonyValue: 250000,
		Status:         clientModels.StatusPending,
	}

	webhookRepo := &mockWebhookRepo{
		getEventByIDFn: func(_ context.Context, _ string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{}, pgx.ErrNoRows
		},
		insertEventFn: func(_ context.Context, eventID, cardID, clientEmail string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{
				EventID:     eventID,
				CardID:      cardID,
				ClientEmail: clientEmail,
				ProcessedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	clientRepo := &mockWebhookClientRepo{
		getByEmailFn: func(_ context.Context, _ string) (*clientModels.ClientResponse, error) {
			return clientWithHighPatrimony, nil
		},
		updateClientFn: func(_ context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error) {
			return &clientModels.ClientResponse{
				Email:    email,
				Status:   status,
				Priority: priority,
			}, nil
		},
	}

	svc := NewWebhookService(webhookRepo, clientRepo, newSuccessfulPipefy())
	resp, err := svc.ProcessEvent(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Priority != clientModels.PriorityHigh {
		t.Errorf("priority: want %q, got %q", clientModels.PriorityHigh, resp.Priority)
	}
	if resp.Status != clientModels.StatusProcessed {
		t.Errorf("status: want %q, got %q", clientModels.StatusProcessed, resp.Status)
	}
}

// TestProcessEvent_NormalPriority verifies that a client with valor_patrimonio < 200.000
// receives prioridade_normal after webhook processing.
func TestProcessEvent_NormalPriority(t *testing.T) {
	req := webhookModels.WebhookEventRequest{
		EventID:     "evt_normal",
		CardID:      "card_789",
		ClientEmail: "maria.lima@example.com",
		Timestamp:   time.Now(),
	}

	clientWithLowPatrimony := &clientModels.ClientResponse{
		ID:             "uuid-2",
		Email:          req.ClientEmail,
		PatrimonyValue: 150000,
		Status:         clientModels.StatusPending,
	}

	webhookRepo := &mockWebhookRepo{
		getEventByIDFn: func(_ context.Context, _ string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{}, pgx.ErrNoRows
		},
		insertEventFn: func(_ context.Context, eventID, cardID, clientEmail string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{
				EventID:     eventID,
				CardID:      cardID,
				ClientEmail: clientEmail,
				ProcessedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	clientRepo := &mockWebhookClientRepo{
		getByEmailFn: func(_ context.Context, _ string) (*clientModels.ClientResponse, error) {
			return clientWithLowPatrimony, nil
		},
		updateClientFn: func(_ context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error) {
			return &clientModels.ClientResponse{
				Email:    email,
				Status:   status,
				Priority: priority,
			}, nil
		},
	}

	svc := NewWebhookService(webhookRepo, clientRepo, newSuccessfulPipefy())
	resp, err := svc.ProcessEvent(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Priority != clientModels.PriorityNormal {
		t.Errorf("priority: want %q, got %q", clientModels.PriorityNormal, resp.Priority)
	}
	if resp.Status != clientModels.StatusProcessed {
		t.Errorf("status: want %q, got %q", clientModels.StatusProcessed, resp.Status)
	}
}

// TestProcessEvent_PriorityBoundary verifies the boundary case where valor_patrimonio == 200.000
// results in prioridade_alta (>= threshold).
func TestProcessEvent_PriorityBoundary(t *testing.T) {
	req := webhookModels.WebhookEventRequest{
		EventID:     "evt_boundary",
		CardID:      "card_999",
		ClientEmail: "boundary@example.com",
		Timestamp:   time.Now(),
	}

	clientAtThreshold := &clientModels.ClientResponse{
		Email:          req.ClientEmail,
		PatrimonyValue: 200000,
		Status:         clientModels.StatusPending,
	}

	webhookRepo := &mockWebhookRepo{
		getEventByIDFn: func(_ context.Context, _ string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{}, pgx.ErrNoRows
		},
		insertEventFn: func(_ context.Context, eventID, cardID, clientEmail string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{
				EventID:     eventID,
				CardID:      cardID,
				ClientEmail: clientEmail,
				ProcessedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	clientRepo := &mockWebhookClientRepo{
		getByEmailFn: func(_ context.Context, _ string) (*clientModels.ClientResponse, error) {
			return clientAtThreshold, nil
		},
		updateClientFn: func(_ context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error) {
			return &clientModels.ClientResponse{Email: email, Status: status, Priority: priority}, nil
		},
	}

	svc := NewWebhookService(webhookRepo, clientRepo, newSuccessfulPipefy())
	resp, err := svc.ProcessEvent(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Priority != clientModels.PriorityHigh {
		t.Errorf("priority at boundary: want %q, got %q", clientModels.PriorityHigh, resp.Priority)
	}
}

// TestProcessEvent_DuplicateEventID verifies that re-sending an already-processed event_id
// is rejected with ErrBadRequest (idempotency guarantee).
func TestProcessEvent_DuplicateEventID(t *testing.T) {
	req := webhookModels.WebhookEventRequest{
		EventID:     "evt_duplicate",
		CardID:      "card_456",
		ClientEmail: "joao.silva@example.com",
		Timestamp:   time.Now(),
	}

	webhookRepo := &mockWebhookRepo{
		getEventByIDFn: func(_ context.Context, eventID string) (webhookdb.WebhookEvent, error) {
			// Simulates event already recorded — returns it without error.
			return webhookdb.WebhookEvent{
				EventID:     eventID,
				CardID:      "card_456",
				ClientEmail: req.ClientEmail,
				ProcessedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	svc := NewWebhookService(webhookRepo, &mockWebhookClientRepo{}, newSuccessfulPipefy())
	_, err := svc.ProcessEvent(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for duplicate event_id, got nil")
	}
	if !errors.Is(err, AppError.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got: %v", err)
	}
}

// TestProcessEvent_ClientNotFound verifies that a webhook for an unknown email returns ErrNotFound.
func TestProcessEvent_ClientNotFound(t *testing.T) {
	req := webhookModels.WebhookEventRequest{
		EventID:     "evt_no_client",
		CardID:      "card_000",
		ClientEmail: "unknown@example.com",
		Timestamp:   time.Now(),
	}

	webhookRepo := &mockWebhookRepo{
		getEventByIDFn: func(_ context.Context, _ string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{}, pgx.ErrNoRows
		},
	}

	clientRepo := &mockWebhookClientRepo{
		getByEmailFn: func(_ context.Context, _ string) (*clientModels.ClientResponse, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewWebhookService(webhookRepo, clientRepo, newSuccessfulPipefy())
	_, err := svc.ProcessEvent(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for unknown client, got nil")
	}
	if !errors.Is(err, AppError.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
