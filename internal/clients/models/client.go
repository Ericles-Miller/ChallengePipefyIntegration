package models

type CreateClientRequest struct {
	Name           string  `json:"cliente_nome"      binding:"required,max=255"`
	Email          string  `json:"cliente_email"     binding:"required,email,max=255"`
	RequestType    string  `json:"tipo_solicitacao"  binding:"required,max=255"`
	PatrimonyValue float64 `json:"valor_patrimonio"  binding:"required,gt=0,lte=9999999999999"`
}

