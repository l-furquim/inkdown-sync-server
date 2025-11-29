# Plano de Projeto - Sistema de Chat em Tempo Real com Go e WebSocket

## 1. Visão Geral do Projeto

Sistema de chat em tempo real construído com Go, utilizando WebSocket (Gorilla) para comunicação bidirecional, REST API para autenticação/cadastro e PouchDB como banco de dados. O projeto seguirá a arquitetura padrão de projetos Go com separação clara de responsabilidades.

## 2. Arquitetura do Projeto

```
chat-system/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point da aplicação
├── internal/
│   ├── config/
│   │   └── config.go              # Configurações da aplicação
│   ├── domain/
│   │   ├── user.go                # Entidade User
│   │   ├── message.go             # Entidade Message
│   │   └── room.go                # Entidade Room
│   ├── repository/
│   │   ├── user_repository.go     # Interface e implementação PouchDB
│   │   └── message_repository.go
│   ├── service/
│   │   ├── auth_service.go        # Lógica de autenticação
│   │   ├── user_service.go        # Lógica de usuários
│   │   └── chat_service.go        # Lógica do chat
│   ├── handler/
│   │   ├── auth_handler.go        # Handlers REST para auth
│   │   ├── user_handler.go        # Handlers REST para users
│   │   └── websocket_handler.go   # Handler WebSocket
│   ├── middleware/
│   │   ├── auth_middleware.go     # Middleware JWT
│   │   ├── cors_middleware.go     # Middleware CORS
│   │   └── logger_middleware.go   # Middleware de logs
│   └── websocket/
│       ├── client.go              # Cliente WebSocket
│       ├── hub.go                 # Hub de conexões
│       └── message.go             # Tipos de mensagens WS
├── pkg/
│   ├── jwt/
│   │   └── jwt.go                 # Utilitários JWT
│   ├── hash/
│   │   └── password.go            # Hash de senhas
│   └── response/
│       └── response.go            # Padronização de respostas
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
└── README.md
```

## 3. Stack Tecnológica

### Backend
- **Linguagem**: Go 1.21+
- **WebSocket**: Gorilla WebSocket
- **Router**: Gorilla Mux
- **Banco de Dados**: PouchDB (via API HTTP)
- **Autenticação**: JWT (golang-jwt/jwt)
- **Hashing**: bcrypt
- **Validação**: go-playground/validator

### DevOps
- **Containerização**: Docker
- **Orquestração**: Docker Compose
- **Proxy Reverso**: Nginx (opcional)

## 4. Endpoints REST API

### Autenticação
```
POST   /api/v1/auth/register    # Cadastro de usuário
POST   /api/v1/auth/login       # Login e geração de JWT
POST   /api/v1/auth/refresh     # Refresh token
POST   /api/v1/auth/logout      # Logout (invalidar token)
```

### Usuários
```
GET    /api/v1/users/me         # Dados do usuário autenticado
PUT    /api/v1/users/me         # Atualizar perfil
GET    /api/v1/users            # Listar usuários (paginado)
```

### WebSocket
```
WS     /ws                       # Conexão WebSocket (requer JWT)
```

## 5. Estrutura de Mensagens WebSocket

### Cliente → Servidor
```json
{
  "type": "message" | "join_room" | "leave_room" | "typing",
  "room_id": "string",
  "content": "string",
  "timestamp": "ISO8601"
}
```

### Servidor → Cliente
```json
{
  "type": "message" | "user_joined" | "user_left" | "typing" | "error",
  "user_id": "string",
  "username": "string",
  "room_id": "string",
  "content": "string",
  "timestamp": "ISO8601"
}
```

## 6. Modelos de Dados

### User
```go
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Message
```go
type Message struct {
    ID        string    `json:"id"`
    RoomID    string    `json:"room_id"`
    UserID    string    `json:"user_id"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Room
```go
type Room struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedBy string    `json:"created_by"`
    CreatedAt time.Time `json:"created_at"`
}
```

## 7. Configuração Docker

### Dockerfile
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### docker-compose.yml
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=pouchdb
      - DB_PORT=5984
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - pouchdb
    networks:
      - chat-network

  pouchdb:
    image: couchdb:3.3
    ports:
      - "5984:5984"
    environment:
      - COUCHDB_USER=admin
      - COUCHDB_PASSWORD=password
    volumes:
      - pouchdb_data:/opt/couchdb/data
    networks:
      - chat-network

volumes:
  pouchdb_data:

networks:
  chat-network:
    driver: bridge
```

## 8. Fluxo de Autenticação

1. **Registro**:
   - Cliente envia POST /api/v1/auth/register
   - Servidor valida dados, faz hash da senha
   - Salva usuário no PouchDB
   - Retorna sucesso

2. **Login**:
   - Cliente envia POST /api/v1/auth/login
   - Servidor valida credenciais
   - Gera JWT token
   - Retorna token + refresh token

3. **Acesso WebSocket**:
   - Cliente conecta com token JWT no header/query
   - Servidor valida token
   - Estabelece conexão WebSocket
   - Adiciona cliente ao Hub

## 9. Gestão de Conexões WebSocket

### Hub Pattern
- **Hub Central**: Gerencia todas as conexões ativas
- **Broadcast**: Envia mensagens para múltiplos clientes
- **Rooms**: Agrupa clientes por salas
- **Heartbeat**: Ping/Pong para detectar conexões mortas

### Client Management
```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    rooms      map[string]map[*Client]bool
}
```

## 10. Segurança

### Medidas Implementadas
- **JWT**: Autenticação stateless
- **Bcrypt**: Hash de senhas (cost 12)
- **CORS**: Configuração restritiva
- **Rate Limiting**: Proteção contra DDoS
- **Input Validation**: Validação de todos os inputs
- **HTTPS**: TLS/SSL em produção
- **Sanitização**: Escape de conteúdo HTML/JS

## 11. Variáveis de Ambiente

```env
# Server
PORT=8080
HOST=0.0.0.0
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5984
DB_USER=admin
DB_PASSWORD=password
DB_NAME=chatdb

# JWT
JWT_SECRET=your-secret-key-here
JWT_EXPIRATION=24h
REFRESH_TOKEN_EXPIRATION=168h

# WebSocket
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_MAX_MESSAGE_SIZE=512
```

## 12. Dependências Go

```go
// go.mod
require (
    github.com/gorilla/mux v1.8.1
    github.com/gorilla/websocket v1.5.1
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/go-playground/validator/v10 v10.16.0
    github.com/joho/godotenv v1.5.1
    golang.org/x/crypto v0.17.0
)
```

## 13. Fases de Desenvolvimento

### Fase 1 - Setup Inicial (Semana 1)
- Configuração do projeto e estrutura de pastas
- Setup Docker e Docker Compose
- Conexão com PouchDB
- Configuração de variáveis de ambiente

### Fase 2 - Autenticação (Semana 2)
- Implementação de registro de usuários
- Sistema de login com JWT
- Middleware de autenticação
- Refresh token

### Fase 3 - WebSocket (Semana 3)
- Setup Gorilla WebSocket
- Implementação do Hub
- Gestão de clientes
- Sistema de broadcast

### Fase 4 - Features de Chat (Semana 4)
- Sistema de salas
- Envio/recebimento de mensagens
- Indicador de digitação
- Histórico de mensagens

### Fase 5 - Testes e Deploy (Semana 5)
- Testes unitários
- Testes de integração
- Testes de carga WebSocket
- Documentação e deploy

## 14. Testes

### Estrutura de Testes
```
├── internal/
│   ├── handler/
│   │   └── auth_handler_test.go
│   ├── service/
│   │   └── auth_service_test.go
│   └── websocket/
│       └── hub_test.go
```

### Ferramentas
- **Testing**: testing package nativo
- **Mocking**: gomock ou testify/mock
- **HTTP Testing**: httptest package
- **Coverage**: go test -cover

## 15. Monitoramento e Logs

- **Structured Logging**: logrus ou zap
- **Metrics**: Prometheus (opcional)
- **Health Check**: Endpoint /health
- **Graceful Shutdown**: Context com timeout


🔒 Segurança e Boas Práticas
Checklist de Segurança

 HTTPS/WSS obrigatório em produção
 Rate limiting por IP e por usuário
 Validação de input
 JWT com expiração curta (15-30 minutos) + refresh tokens
 Hash de senhas
 Criptografia E2E - servidor nunca vê conteúdo decriptado
 Auditoria de acesso - log de todas as operações
 CORS configurado corretamente
 SQL Injection prevenido
 XSS prevenido (sanitização de inputs)

Performance

 Compressão de WebSocket (permessage-deflate)
 Batching de mensagens pequenas
 Lazy loading de histórico de versões
 Paginação de listas grandes
 Índices de banco em colunas frequentemente consultadas
 Pooling de conexões de banco

Este plano fornece uma base sólida para desenvolver um sistema de chat completo e escalável usando Go, seguindo as melhores práticas da comunidade Go e patterns de mercado.