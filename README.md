# Inkdown Sync Server

Servidor de autenticação, cadastro e sincronização de notas markdown para o [Inkdown](https://github.com/inkdown), construído com Go seguindo boas práticas de desenvolvimento.

## 🛠️ Stack Tecnológica

- Go lang
- Gorilla Mux
- Gorilla WebSocket
- JWT (golang-jwt/jwt/v5)
- CouchDB 3.3
- go-playground/validator
- bcrypt

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

### Notas

```
GET    /api/v1/notes            # Listar todas as notas do usuário
GET    /api/v1/notes/:id        # Buscar nota por ID
POST   /api/v1/notes            # Criar nova nota
PUT    /api/v1/notes/:id        # Atualizar nota existente
DELETE /api/v1/notes/:id        # Deletar nota (soft delete)
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
JWT_SECRET=your-secret-key-here  # MUDE EM PRODUÇÃO!
JWT_EXPIRATION=15m

# WebSocket
WS_MAX_MESSAGE_SIZE=10485760  # 10MB
```

## Estrutura do Projeto

```
inkdown-sync-server/
├── cmd/server/           # Entry point da aplicação
├── internal/
│   ├── config/           # Configurações
│   ├── domain/           # Entidades de domínio
│   ├── repository/       # Camada de dados
│   ├── service/          # Lógica de negócio
│   ├── handler/          # HTTP/WebSocket handlers
│   ├── middleware/       # Middlewares (auth, cors, logging)
│   └── websocket/        # Infraestrutura WebSocket
├── pkg/                  # Pacotes reutilizáveis
│   ├── jwt/              # Utilitários JWT
│   ├── hash/             # Hash de senhas
│   ├── response/         # Respostas padronizadas
│   └── validator/        # Validadores customizados
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

## Licença

MIT

## Contribuindo

Contribuições são bem-vindas! Por favor, abra uma issue ou PR.

---

**Desenvolvido para o Inkdown** - Um editor de notas markdown multiplataforma moderno e eficiente.