package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/models"
	AppError "github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/appError"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/pipefy"
	"github.com/jackc/pgx/v5"
)

// --- mock implementations ---

type mockClientRepo struct {
	createFn       func(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error)
	getByEmailFn   func(ctx context.Context, email string) (*models.ClientResponse, error)
	updateClientFn func(ctx context.Context, email string, status models.ClientStatus, priority models.ClientPriority) (*models.ClientResponse, error)
}

func (m *mockClientRepo) Create(ctx context.Context, req models.CreateClientRequest) (*models.ClientResponse, error) {
	return m.createFn(ctx, req)
}

func (m *mockClientRepo) GetByEmail(ctx context.Context, email string) (*models.ClientResponse, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockClientRepo) UpdateClient(ctx context.Context, email string, status models.ClientStatus, priority models.ClientPriority) (*models.ClientResponse, error) {
	return m.updateClientFn(ctx, email, status, priority)
}

type mockPipefyClient struct {
	createCardFn          func(ctx context.Context, name, email, requestType string, patrimony float64) (*pipefy.CardResult, error)
	moveCardToProcessedFn func(ctx context.Context, cardID string) error
	updateCardFieldFn     func(ctx context.Context, cardID, fieldID, value string) error
}

func (m *mockPipefyClient) CreateCard(ctx context.Context, name, email, requestType string, patrimony float64) (*pipefy.CardResult, error) {
	return m.createCardFn(ctx, name, email, requestType, patrimony)
}

func (m *mockPipefyClient) MoveCardToProcessed(ctx context.Context, cardID string) error {
	return m.moveCardToProcessedFn(ctx, cardID)
}

func (m *mockPipefyClient) UpdateCardField(ctx context.Context, cardID, fieldID, value string) error {
	return m.updateCardFieldFn(ctx, cardID, fieldID, value)
}

// --- tests ---

// TestCreateClient_Success verifies that a client is created and persisted with status "Aguardando Análise".
func TestCreateClient_Success(t *testing.T) {
	req := models.CreateClientRequest{
		Name:           "João Silva",
		Email:          "joao.silva@example.com",
		RequestType:    "Atualização cadastral",
		PatrimonyValue: 250000,
	}

	saved := &models.ClientResponse{
		ID:             "abc-uuid",
		Name:           req.Name,
		Email:          req.Email,
		RequestType:    req.RequestType,
		PatrimonyValue: req.PatrimonyValue,
		Status:         models.StatusPending,
	}

	repo := &mockClientRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.ClientResponse, error) {
			return nil, pgx.ErrNoRows
		},
		createFn: func(_ context.Context, r models.CreateClientRequest) (*models.ClientResponse, error) {
			return saved, nil
		},
	}

	pipefyClient := &mockPipefyClient{
		createCardFn: func(_ context.Context, _, _, _ string, _ float64) (*pipefy.CardResult, error) {
			return &pipefy.CardResult{ID: "card-1", Title: req.Name}, nil
		},
	}

	svc := NewClientService(repo, pipefyClient)
	got, err := svc.CreateClient(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got == nil {
		t.Fatal("expected a ClientResponse, got nil")
	}
	if got.Email != req.Email {
		t.Errorf("email: want %q, got %q", req.Email, got.Email)
	}
	if got.Status != models.StatusPending {
		t.Errorf("status: want %q, got %q", models.StatusPending, got.Status)
	}
	if got.PatrimonyValue != req.PatrimonyValue {
		t.Errorf("patrimony: want %v, got %v", req.PatrimonyValue, got.PatrimonyValue)
	}
}

// TestCreateClient_DuplicateEmail verifies that creating a client with an already-registered email returns ErrBadRequest.
func TestCreateClient_DuplicateEmail(t *testing.T) {
	req := models.CreateClientRequest{
		Name:           "João Silva",
		Email:          "joao.silva@example.com",
		RequestType:    "Atualização cadastral",
		PatrimonyValue: 250000,
	}

	repo := &mockClientRepo{
		getByEmailFn: func(_ context.Context, email string) (*models.ClientResponse, error) {
			return &models.ClientResponse{Email: email}, nil
		},
	}

	svc := NewClientService(repo, &mockPipefyClient{})
	_, err := svc.CreateClient(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	if !errors.Is(err, AppError.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got: %v", err)
	}
}

// TestCreateClient_PipefyError verifies that a Pipefy failure is propagated as ErrInternalServer.
func TestCreateClient_PipefyError(t *testing.T) {
	req := models.CreateClientRequest{
		Name:           "João Silva",
		Email:          "joao.silva@example.com",
		RequestType:    "Atualização cadastral",
		PatrimonyValue: 100000,
	}

	repo := &mockClientRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.ClientResponse, error) {
			return nil, pgx.ErrNoRows
		},
		createFn: func(_ context.Context, _ models.CreateClientRequest) (*models.ClientResponse, error) {
			return &models.ClientResponse{Email: req.Email, Status: models.StatusPending}, nil
		},
	}

	pipefyClient := &mockPipefyClient{
		createCardFn: func(_ context.Context, _, _, _ string, _ float64) (*pipefy.CardResult, error) {
			return nil, errors.New("pipefy unavailable")
		},
	}

	svc := NewClientService(repo, pipefyClient)
	_, err := svc.CreateClient(context.Background(), req)

	if err == nil {
		t.Fatal("expected error from Pipefy failure, got nil")
	}
	if !errors.Is(err, AppError.ErrInternalServer) {
		t.Errorf("expected ErrInternalServer, got: %v", err)
	}
}
