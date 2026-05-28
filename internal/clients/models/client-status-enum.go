package models

type ClientStatus string

const (
	StatusPending   ClientStatus = "Aguardando Análise"
	StatusProcessed ClientStatus = "Processado"
)


