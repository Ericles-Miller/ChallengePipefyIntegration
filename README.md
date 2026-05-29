# ChallengePipefyIntegration

API em Go para gerenciamento de clientes e integração com o Pipefy via GraphQL.

## Tecnologias

- **Go** — linguagem principal
- **Gin** — framework HTTP
- **PostgreSQL** — banco de dados relacional
- **sqlc** — geração de código type-safe a partir de SQL
- **Goose** — migrations de banco de dados
- **Docker / Docker Compose** — infraestrutura local

---

## Estrutura de Pastas

```
.
├── api/                        # Inicialização do servidor e injeção de dependências
├── internal/
│   ├── clients/                # Domínio de clientes
│   │   ├── models/             # Structs de request, response e enums
│   │   ├── repositories/       # Acesso ao banco de dados
│   │   ├── services/           # Regras de negócio + testes unitários
│   │   └── db/                 # Código gerado pelo sqlc
│   ├── webhooks/               # Domínio de webhooks
│   │   ├── models/
│   │   ├── repositories/
│   │   ├── services/           # Regras de negócio + testes unitários
│   │   └── db/
│   └── database/               # Conexão e migrations
├── pkg/
│   ├── pipefy/                 # Client GraphQL do Pipefy (mutations)
│   └── appError/               # Erros customizados com tipagem
└── main.go
```

---

## Pré-requisitos

- [Go 1.21+](https://golang.org/dl/)
- [Docker](https://www.docker.com/) e Docker Compose

---

## Execução Local

### 1. Clone o repositório

```bash
git clone https://github.com/Ericles-Miller/ChallengePipefyIntegration.git
cd ChallengePipefyIntegration
```

### 2. Configure as variáveis de ambiente

Crie um arquivo `.env` na raiz do projeto:

```env
PORT=8080
DATABASE_URL=postgres://myuser:mysecretpassword@localhost:5432/challenge_pipefy

TOKEN_PIPEFY=seu_token_aqui
PIPEFY_API_URL=https://api.pipefy.com/graphql
PIPEFY_PIPE_ID=seu_pipe_id_aqui
PIPEFY_PHASE_PROCESSED=seu_phase_id_aqui
```

> As credenciais de banco (`myuser`, `mysecretpassword`, `challenge_pipefy`) já estão configuradas no `docker-compose.yml`.

### 3. Suba o banco de dados

```bash
docker-compose up -d
```

### 4. Instale as dependências

```bash
go mod download
```

### 5. Inicie a API

```bash
go run main.go
```

As migrations são executadas automaticamente na inicialização. A API estará disponível em `http://localhost:8080`.

---

## Endpoints

### `GET /health`

Verifica se a API está no ar.

```bash
curl http://localhost:8080/health
```

### `POST /clientes`

Cria um novo cliente e mapeia um card no Pipefy.

```bash
curl -X POST http://localhost:8080/clientes \
  -H "Content-Type: application/json" \
  -d '{
    "cliente_nome": "João Silva",
    "cliente_email": "joao.silva@example.com",
    "tipo_solicitacao": "Atualização cadastral",
    "valor_patrimonio": 250000
  }'
```

**Resposta (201 Created):**

```json
{
  "data": {
    "id": "uuid-gerado",
    "cliente_nome": "João Silva",
    "cliente_email": "joao.silva@example.com",
    "tipo_solicitacao": "Atualização cadastral",
    "valor_patrimonio": 250000,
    "status": "Aguardando Análise"
  }
}
```

---

### `POST /webhooks/pipefy/card-updated`

Simula o recebimento de um evento do Pipefy quando um card é atualizado.

```bash
curl -X POST http://localhost:8080/webhooks/pipefy/card-updated \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt_123",
    "card_id": "card_456",
    "cliente_email": "joao.silva@example.com",
    "timestamp": "2026-05-18T12:00:00Z"
  }'
```

**Resposta (200 OK):**

```json
{
  "data": {
    "event_id": "evt_123",
    "cliente_email": "joao.silva@example.com",
    "status": "Processado",
    "prioridade": "prioridade_alta",
    "processed_at": "2026-05-28T10:00:00Z"
  }
}
```

> A prioridade é calculada automaticamente: `valor_patrimonio >= 200.000` → `prioridade_alta`, caso contrário → `prioridade_normal`.

---

## Documentação Swagger

Com a API rodando, acesse:

```
http://localhost:8080/swagger/index.html
```

---

## Testes

Os testes são unitários, não dependem de banco de dados ou variáveis de ambiente — podem ser executados diretamente.

### Rodar todos os testes

```bash
go test ./...
```

### Rodar apenas os testes de serviço (com output detalhado)

```bash
go test ./internal/clients/services/... ./internal/webhooks/services/... -v
```

### Rodar um teste específico

```bash
go test ./internal/webhooks/services/... -run TestProcessEvent_DuplicateEventID -v
```

### Cenários cobertos

| Teste | Descrição |
|---|---|
| `TestCreateClient_Success` | Criação de cliente com payload válido |
| `TestCreateClient_DuplicateEmail` | Bloqueio de e-mail duplicado |
| `TestCreateClient_PipefyError` | Falha no Pipefy retorna erro interno |
| `TestProcessEvent_HighPriority` | Patrimônio >= 200.000 → `prioridade_alta` |
| `TestProcessEvent_NormalPriority` | Patrimônio < 200.000 → `prioridade_normal` |
| `TestProcessEvent_PriorityBoundary` | Patrimônio exatamente 200.000 → `prioridade_alta` |
| `TestProcessEvent_DuplicateEventID` | Bloqueio de `event_id` já processado (idempotência) |
| `TestProcessEvent_ClientNotFound` | Webhook para e-mail inexistente retorna 404 |

---

## Integração com o Pipefy (GraphQL)

As mutations GraphQL estão centralizadas em [`pkg/pipefy/client.go`](pkg/pipefy/client.go).

**Criação de card (`createCard`)** — acionada ao criar um cliente:

```graphql
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
}
```

**Mover card de fase (`moveCardToPhase`)** — acionada ao processar o webhook:

```graphql
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
}
```

**Atualizar campo do card (`updateCardField`)** — atualiza a prioridade no Pipefy:

```graphql
mutation UpdateCardField($cardId: ID!, $fieldId: ID!, $newValue: [UndefinedInput]) {
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
}
```

---

## Visão de Produção (AWS)

Em um ambiente de produção na AWS, a arquitetura escalaria da seguinte forma:

### Camada de entrada
- **API Gateway** expõe os dois endpoints (`POST /clientes` e `POST /webhooks/pipefy/card-updated`) de forma gerenciada, com autenticação, rate limiting e logging automáticos.

### Processamento
- **AWS Lambda** substitui o servidor Go sempre ativo. Cada endpoint vira uma função independente, escalando automaticamente conforme a demanda e sem custo em idle.
- Para o webhook especificamente, o API Gateway pode entregar o evento para uma fila **SQS** antes do Lambda. Isso garante que, mesmo com pico de chamadas do Pipefy, nenhum evento seja perdido, e o Lambda processa na velocidade que conseguir (desacoplamento e resiliência).

### Persistência
- **RDS (PostgreSQL)** para manter a mesma estrutura relacional atual, com Multi-AZ para alta disponibilidade.
- Alternativamente, **DynamoDB** para a tabela de `webhook_events` (chave primária = `event_id`), aproveitando a verificação de idempotência em O(1) com throughput praticamente ilimitado.

### Idempotência em escala
- Com múltiplas instâncias Lambda processando webhooks em paralelo, a idempotência é garantida pelo `event_id` como chave primária no banco — o banco rejeita inserções duplicadas mesmo sob concorrência.

```
Pipefy → API Gateway → SQS → Lambda (ProcessEvent) → RDS / DynamoDB
Cliente → API Gateway → Lambda (CreateClient)       → RDS
                                                     ↕
                                               Pipefy GraphQL API
```
