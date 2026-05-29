package services

import (
	"context"
	"errors"
	"testing"
	"time"

	clientModels "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/models"
	clientRepositories "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/repositories"
	webhookdb "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/db"
	webhookModels "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/models"
	webhookRepositories "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/repositories"
	AppError "github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/appError"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/pipefy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- mock transaction ---

type noopTx struct{ pgx.Tx }

func (t *noopTx) Commit(_ context.Context) error   { return nil }
func (t *noopTx) Rollback(_ context.Context) error { return nil }

type mockTxBeginner struct{}

func (m *mockTxBeginner) Begin(_ context.Context) (pgx.Tx, error) {
	return &noopTx{}, nil
}

// --- mock repositories ---

type mockWebhookClientRepo struct {
	getByEmailFn   func(ctx context.Context, email string) (*clientModels.ClientResponse, error)
	updateClientFn func(ctx context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error)
}

func (m *mockWebhookClientRepo) Create(_ context.Context, _ clientModels.CreateClientRequest) (*clientModels.ClientResponse, error) {
	panic("Create not expected in webhook service tests")
}

func (m *mockWebhookClientRepo) GetByEmail(ctx context.Context, email string) (*clientModels.ClientResponse, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockWebhookClientRepo) UpdateClient(ctx context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error) {
	return m.updateClientFn(ctx, email, status, priority)
}

func (m *mockWebhookClientRepo) WithTx(_ pgx.Tx) clientRepositories.ClientRepository {
	return m
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

func (m *mockWebhookRepo) WithTx(_ pgx.Tx) webhookRepositories.WebhookRepository {
	return m
}

// --- mock pipefy ---

type mockWebhookPipefy struct {
	moveCardToProcessedFn func(ctx context.Context, cardID string) error
	updateCardFieldFn     func(ctx context.Context, cardID, fieldID, value string) error
}

func (m *mockWebhookPipefy) CreateCard(_ context.Context, _, _, _ string, _ float64) (*pipefy.CardResult, error) {
	panic("CreateCard not expected in webhook service tests")
}

func (m *mockWebhookPipefy) MoveCardToProcessed(ctx context.Context, cardID string) error {
	return m.moveCardToProcessedFn(ctx, cardID)
}

func (m *mockWebhookPipefy) UpdateCardField(ctx context.Context, cardID, fieldID, value string) error {
	return m.updateCardFieldFn(ctx, cardID, fieldID, value)
}

// --- helpers ---

func newSuccessfulPipefy() *mockWebhookPipefy {
	return &mockWebhookPipefy{
		moveCardToProcessedFn: func(_ context.Context, _ string) error { return nil },
		updateCardFieldFn:     func(_ context.Context, _, _, _ string) error { return nil },
	}
}

func newWebhookService(wRepo *mockWebhookRepo, cRepo *mockWebhookClientRepo, p *mockWebhookPipefy) WebhookService {
	return &webhookService{webhookRepo: wRepo, clientRepo: cRepo, pipefy: p, db: &mockTxBeginner{}}
}

func stubInsertEvent(eventID, cardID, clientEmail string) func(context.Context, string, string, string) (webhookdb.WebhookEvent, error) {
	return func(_ context.Context, id, card, email string) (webhookdb.WebhookEvent, error) {
		return webhookdb.WebhookEvent{
			EventID:     id,
			CardID:      card,
			ClientEmail: email,
			ProcessedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil
	}
}

// --- tests ---

// TestProcessEvent_HighPriority verifica que valor_patrimonio >= 200.000 resulta em prioridade_alta.
func TestProcessEvent_HighPriority(t *testing.T) {
	req := webhookModels.WebhookEventRequest{
		EventID:     "evt_high",
		CardID:      "card_456",
		ClientEmail: "joao.silva@example.com",
		Timestamp:   time.Now(),
	}

	webhookRepo := &mockWebhookRepo{
		getEventByIDFn: func(_ context.Context, _ string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{}, pgx.ErrNoRows
		},
		insertEventFn: stubInsertEvent(req.EventID, req.CardID, req.ClientEmail),
	}

	clientRepo := &mockWebhookClientRepo{
		getByEmailFn: func(_ context.Context, _ string) (*clientModels.ClientResponse, error) {
			return &clientModels.ClientResponse{Email: req.ClientEmail, PatrimonyValue: 250000}, nil
		},
		updateClientFn: func(_ context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error) {
			return &clientModels.ClientResponse{Email: email, Status: status, Priority: priority}, nil
		},
	}

	svc := newWebhookService(webhookRepo, clientRepo, newSuccessfulPipefy())
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

// TestProcessEvent_NormalPriority verifica que valor_patrimonio < 200.000 resulta em prioridade_normal.
func TestProcessEvent_NormalPriority(t *testing.T) {
	req := webhookModels.WebhookEventRequest{
		EventID:     "evt_normal",
		CardID:      "card_789",
		ClientEmail: "maria.lima@example.com",
		Timestamp:   time.Now(),
	}

	webhookRepo := &mockWebhookRepo{
		getEventByIDFn: func(_ context.Context, _ string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{}, pgx.ErrNoRows
		},
		insertEventFn: stubInsertEvent(req.EventID, req.CardID, req.ClientEmail),
	}

	clientRepo := &mockWebhookClientRepo{
		getByEmailFn: func(_ context.Context, _ string) (*clientModels.ClientResponse, error) {
			return &clientModels.ClientResponse{Email: req.ClientEmail, PatrimonyValue: 150000}, nil
		},
		updateClientFn: func(_ context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error) {
			return &clientModels.ClientResponse{Email: email, Status: status, Priority: priority}, nil
		},
	}

	svc := newWebhookService(webhookRepo, clientRepo, newSuccessfulPipefy())
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

// TestProcessEvent_PriorityBoundary verifica que valor_patrimonio exatamente 200.000 resulta em prioridade_alta.
func TestProcessEvent_PriorityBoundary(t *testing.T) {
	req := webhookModels.WebhookEventRequest{
		EventID:     "evt_boundary",
		CardID:      "card_999",
		ClientEmail: "boundary@example.com",
		Timestamp:   time.Now(),
	}

	webhookRepo := &mockWebhookRepo{
		getEventByIDFn: func(_ context.Context, _ string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{}, pgx.ErrNoRows
		},
		insertEventFn: stubInsertEvent(req.EventID, req.CardID, req.ClientEmail),
	}

	clientRepo := &mockWebhookClientRepo{
		getByEmailFn: func(_ context.Context, _ string) (*clientModels.ClientResponse, error) {
			return &clientModels.ClientResponse{Email: req.ClientEmail, PatrimonyValue: 200000}, nil
		},
		updateClientFn: func(_ context.Context, email string, status clientModels.ClientStatus, priority clientModels.ClientPriority) (*clientModels.ClientResponse, error) {
			return &clientModels.ClientResponse{Email: email, Status: status, Priority: priority}, nil
		},
	}

	svc := newWebhookService(webhookRepo, clientRepo, newSuccessfulPipefy())
	resp, err := svc.ProcessEvent(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Priority != clientModels.PriorityHigh {
		t.Errorf("priority at boundary: want %q, got %q", clientModels.PriorityHigh, resp.Priority)
	}
}

// TestProcessEvent_DuplicateEventID verifica que reenviar um event_id já processado é bloqueado (idempotência).
func TestProcessEvent_DuplicateEventID(t *testing.T) {
	req := webhookModels.WebhookEventRequest{
		EventID:     "evt_duplicate",
		CardID:      "card_456",
		ClientEmail: "joao.silva@example.com",
		Timestamp:   time.Now(),
	}

	webhookRepo := &mockWebhookRepo{
		getEventByIDFn: func(_ context.Context, eventID string) (webhookdb.WebhookEvent, error) {
			return webhookdb.WebhookEvent{
				EventID:     eventID,
				ProcessedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	svc := newWebhookService(webhookRepo, &mockWebhookClientRepo{}, newSuccessfulPipefy())
	_, err := svc.ProcessEvent(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for duplicate event_id, got nil")
	}
	if !errors.Is(err, AppError.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got: %v", err)
	}
}

// TestProcessEvent_ClientNotFound verifica que webhook para e-mail inexistente retorna ErrNotFound.
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

	svc := newWebhookService(webhookRepo, clientRepo, newSuccessfulPipefy())
	_, err := svc.ProcessEvent(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for unknown client, got nil")
	}
	if !errors.Is(err, AppError.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
