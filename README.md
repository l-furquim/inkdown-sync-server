# Notes Sync Server

Api de autenticação, cadastro e sincronização de notas entre diversos dispositivos, utilizando websocket para
sincronização em tempo real, CouchDB como banco de dados NoSQL e testes unitarios para garantir a qualidade
e funcionalidade do codigo. Contem encriptografia e2e do conteudo e titulo das notas, garantindo privacidade
e segurança da sincronização.

## Instalação e Configuração

### 1. Clone o Repositório

```bash
git clone <repository-url>
cd inkdown-sync-server
```

### 2. Configure as Variáveis de Ambiente

```bash
cp .env.example .env
# Edite o .env e altere JWT_SECRET e outras configurações conforme necessário
```

### 3. Opção A: Execução com Docker Compose (Recomendado)

```bash
# Build e start de todos os serviços (app + CouchDB)
docker-compose up --build

# O servidor estará disponível em http://localhost:8080
# CouchDB Admin UI em http://localhost:5984/_utils
```

### 3. Opção B: Execução Local

```bash
# Instalar dependências
go mod download

# Build
go build -o server cmd/server/main.go

# Executar (certifique-se de que o CouchDB está rodando)
./server
```

## Endpoints da API

### Autenticação

```
POST   /api/v1/auth/register    # Cadastro de novo usuário
POST   /api/v1/auth/login       # Login (retorna JWT + refresh token)
POST   /api/v1/auth/refresh     # Renovar access token
POST   /api/v1/auth/logout      # Logout (invalidar tokens)
```

### Usuários

```
GET    /api/v1/users/me         # Obter dados do usuário autenticado
PUT    /api/v1/users/me         # Atualizar perfil
```

### Dispositivos

```
POST   /api/v1/devices/register # Registra um novo dispositivo
GET    /api/v1/devices          # Lista dispositivos ativos
DELETE /api/v1/devices/{id}     # Revoga acesso de um dispositivo
```

### Segurança (E2EE)

```
POST   /api/v1/security/keys/setup # Upload da chave mestra criptografada
GET    /api/v1/security/keys/sync  # Download da chave mestra criptografada
```

### Notas

```
POST   /api/v1/notes            # Criar nova nota ou diretório
GET    /api/v1/notes            # Listar todas as notas
GET    /api/v1/notes/{id}       # Obter detalhes de uma nota
PUT    /api/v1/notes/{id}       # Atualizar uma nota
DELETE /api/v1/notes/{id}       # Deletar uma nota (soft delete)
```

### WebSocket

```
WS     /ws?token=<jwt>          # Conexão para sincronização em tempo real
```

### Health Check

```
GET    /health                  # Verificar status do servidor
```

## Testes

```bash
# Executar todos os testes
go test ./... -v

# Testes com coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Testes de integração (requer Docker)
docker-compose up -d couchdb
go test ./internal/handler -tags=integration -v
```

## Build para Produção

```bash
# Build otimizado
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Ou usar Docker
docker build -t inkdown-sync-server:latest .
```

## Variáveis de Ambiente

Veja `.env.example` para a lista completa. Principais:

```env
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5984
DB_NAME=inkdown

# JWT
JWT_SECRET=your-secret-key-here 
JWT_EXPIRATION=15m

# WebSocket
WS_MAX_MESSAGE_SIZE=10485760  # 10MB
```
