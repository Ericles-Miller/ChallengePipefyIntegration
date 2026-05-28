package models

type CreateClientRequest struct {
	Name            string  `json:"cliente_nome" binding:"required"`
	Email           string  `json:"cliente_email" binding:"required,email"`
	RequestType     string  `json:"tipo_solicitacao" binding:"required"`
	PatrimonyValue  float64 `json:"valor_patrimonio" binding:"required,gt=0"`
}

