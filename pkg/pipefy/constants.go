package pipefy

import "errors"

var ErrUnauthorized = errors.New("pipefy: unauthorized")

const (
	FieldClientName  = "requester_name"
	FieldClientEmail = "contact_email"
	FieldRequestType = "tipo_solicitacao"
	FieldPatrimony   = "num_rico"
	FieldPriority    = "prioridade"
)
