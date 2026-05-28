package models

import (
	clientdb "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/db"
)

func ToResponse(c clientdb.Client) *ClientResponse {
	return &ClientResponse{
		ID:             c.ID.String(),
		Name:           c.Name,
		Email:          c.Email,
		RequestType:    c.RequestType,
		PatrimonyValue: c.PatrimonyValue,
		Status:         ClientStatus(c.Status),
		Priority:       ClientPriority(c.Priority.String),
	}
}