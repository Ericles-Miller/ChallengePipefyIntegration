package models

type ClientResponse struct {
	ID             string         `json:"id"`
	Name           string         `json:"cliente_nome"`
	Email          string         `json:"cliente_email"`
	RequestType    string         `json:"tipo_solicitacao"`
	PatrimonyValue float64        `json:"valor_patrimonio"`
	Status         ClientStatus   `json:"status"`
	Priority       ClientPriority `json:"prioridade,omitempty"`
}
