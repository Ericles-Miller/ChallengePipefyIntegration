package repositories

import (
	"context"

	webhookdb "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookRepository interface {
	InsertEvent(ctx context.Context, eventID, cardID, clientEmail string) (webhookdb.WebhookEvent, error)
	GetEventByID(ctx context.Context, eventID string) (webhookdb.WebhookEvent, error)
	WithTx(tx pgx.Tx) WebhookRepository
}

type webhookRepository struct {
	queries *webhookdb.Queries
}

func NewWebhookRepository(pool *pgxpool.Pool) WebhookRepository {
	return &webhookRepository{queries: webhookdb.New(pool)}
}

func (r *webhookRepository) InsertEvent(ctx context.Context, eventID, cardID, clientEmail string) (webhookdb.WebhookEvent, error) {
	return r.queries.InsertWebhookEvent(ctx, webhookdb.InsertWebhookEventParams{
		EventID:     eventID,
		CardID:      cardID,
		ClientEmail: clientEmail,
	})
}

func (r *webhookRepository) GetEventByID(ctx context.Context, eventID string) (webhookdb.WebhookEvent, error) {
	return r.queries.GetWebhookEventByID(ctx, eventID)
}

func (r *webhookRepository) WithTx(tx pgx.Tx) WebhookRepository {
	return &webhookRepository{queries: r.queries.WithTx(tx)}
}
