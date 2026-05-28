package pipefy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)


const mutationCreateCard = `
mutation CreateCard($pipeId: ID!, $fields: [FieldValueInput]) {
  createCard(input: {
    pipe_id: $pipeId
    fields_attributes: $fields
  }) {
    card {
      id
      title
    }
  }
}`

const mutationMoveCardToPhase = `
mutation MoveCardToPhase($cardId: ID!, $phaseId: ID!) {
  moveCardToPhase(input: {
    card_id: $cardId
    destination_phase_id: $phaseId
  }) {
    card {
      id
      current_phase {
        id
        name
      }
    }
  }
}`

const mutationUpdateCardField = `
mutation UpdateCardField($cardId: ID!, $fieldId: String!, $newValue: [UndefinedInput]) {
  updateCardField(input: {
    card_id: $cardId
    field_id: $fieldId
    new_value: $newValue
  }) {
    card {
      id
    }
    success
  }
}`

type PipefyClient interface {
	CreateCard(ctx context.Context, name, email, requestType string, patrimony float64) (*CardResult, error)
	MoveCardToProcessed(ctx context.Context, cardID string) error
	UpdateCardField(ctx context.Context, cardID, fieldID, value string) error
}

type pipefyClient struct {
	httpClient     *http.Client
	token          string
	apiURL         string
	pipeID         string
	phaseProcessed string
}

func NewPipefyClient() PipefyClient {
	return &pipefyClient{
		httpClient:     &http.Client{},
		token:          os.Getenv("TOKEN_PIPEFY"),
		apiURL:         os.Getenv("PIPEFY_API_URL"),
		pipeID:         os.Getenv("PIPEFY_PIPE_ID"),
		phaseProcessed: os.Getenv("PIPEFY_PHASE_PROCESSED"),
	}
}

func (c *pipefyClient) CreateCard(ctx context.Context, name, email, requestType string, patrimony float64) (*CardResult, error) {
	variables := map[string]any{
		"pipeId": c.pipeID,
		"fields": []map[string]any{
			{"field_id": FieldClientName, "field_value": name},
			{"field_id": FieldClientEmail, "field_value": email},
			{"field_id": FieldRequestType, "field_value": requestType},
			{"field_id": FieldPatrimony, "field_value": fmt.Sprintf("%g", patrimony)},
		},
	}

	var result graphQLResponse[createCardData]
	if err := c.do(ctx, mutationCreateCard, variables, &result); err != nil {
		return nil, err
	}

	card := result.Data.CreateCard.Card
	return &card, nil
}

func (c *pipefyClient) MoveCardToProcessed(ctx context.Context, cardID string) error {
	variables := map[string]any{
		"cardId":  cardID,
		"phaseId": c.phaseProcessed,
	}

	var result graphQLResponse[moveCardToPhaseData]
	return c.do(ctx, mutationMoveCardToPhase, variables, &result)
}

func (c *pipefyClient) UpdateCardField(ctx context.Context, cardID, fieldID, value string) error {
	variables := map[string]any{
		"cardId":   cardID,
		"fieldId":  fieldID,
		"newValue": value,
	}

	var result graphQLResponse[updateCardFieldData]
	return c.do(ctx, mutationUpdateCardField, variables, &result)
}

func (c *pipefyClient) do(ctx context.Context, query string, variables map[string]any, out any) error {
	body, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("pipefy: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("pipefy: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pipefy: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("pipefy: failed to read response: %w", err)
	}

	var errCheck struct {
		Errors []graphQLError `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &errCheck); err == nil && len(errCheck.Errors) > 0 {
		return fmt.Errorf("pipefy: %s", errCheck.Errors[0].Message)
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("pipefy: failed to parse response: %w", err)
	}

	return nil
}
