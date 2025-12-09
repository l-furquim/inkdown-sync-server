# Documentação da API Inkdown Sync Server

Esta documentação detalha todos os endpoints disponíveis na API, incluindo exemplos de requisição e resposta.

## Autenticação

### Registrar Usuário
Cria uma nova conta de usuário.

**Endpoint:** `POST /api/v1/auth/register`

**Body:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePassword123!"
}
```

**Response (201 Created):**
```json
{
  "message": "User registered successfully. Please login."
}
```

### Login
Autentica um usuário e retorna tokens de acesso.

**Endpoint:** `POST /api/v1/auth/login`

**Body:**
```json
{
  "email": "john@example.com",
  "password": "SecurePassword123!"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1Ni...",
  "refresh_token": "eyJhbGciOiJIUzI1Ni...",
  "expires_in": 900,
  "user": {
    "id": "user:uuid...",
    "username": "johndoe",
    "email": "john@example.com"
  }
}
```

### Refresh Token
Renova o token de acesso usando um refresh token válido.

**Endpoint:** `POST /api/v1/auth/refresh`

**Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1Ni..."
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1Ni...",
  "refresh_token": "eyJhbGciOiJIUzI1Ni...",
  "expires_in": 900
}
```

### Logout
Invalida a sessão (no lado do cliente, remove-se o token).

**Endpoint:** `POST /api/v1/auth/logout`

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

---

## Usuários

### Obter Perfil
Retorna os dados do usuário autenticado.

**Endpoint:** `GET /api/v1/users/me`

**Response (200 OK):**
```json
{
  "id": "user:uuid...",
  "username": "johndoe",
  "email": "john@example.com",
  "created_at": "2023-11-28T10:00:00Z",
  "updated_at": "2023-11-28T10:00:00Z"
}
```

### Atualizar Perfil
Atualiza informações do usuário (ex: username).

**Endpoint:** `PUT /api/v1/users/me`

**Body:**
```json
{
  "username": "john_updated"
}
```

**Response (200 OK):**
```json
{
  "id": "user:uuid...",
  "username": "john_updated",
  "email": "john@example.com",
  "created_at": "2023-11-28T10:00:00Z",
  "updated_at": "2023-11-28T12:00:00Z"
}
```

---

## Dispositivos

### Registrar Dispositivo
Registra o dispositivo atual para rastreamento e controle.

**Endpoint:** `POST /api/v1/devices/register`

**Body:**
```json
{
  "name": "MacBook Pro de John",
  "type": "desktop",
  "os": "darwin",
  "app_version": "1.0.0"
}
```

**Response (201 Created):**
```json
{
  "id": "device:uuid...",
  "name": "MacBook Pro de John",
  "type": "desktop",
  "os": "darwin",
  "last_active": "2023-11-28T12:00:00Z",
  "is_revoked": false
}
```

### Listar Dispositivos
Lista todos os dispositivos registrados na conta.

**Endpoint:** `GET /api/v1/devices`

**Response (200 OK):**
```json
[
  {
    "id": "device:uuid...",
    "name": "MacBook Pro de John",
    "type": "desktop",
    "os": "darwin",
    "last_active": "2023-11-28T12:00:00Z",
    "is_revoked": false
  }
]
```

### Revogar Dispositivo
Remove o acesso de um dispositivo específico.

**Endpoint:** `DELETE /api/v1/devices/{id}`

**Response (200 OK):**
```json
{
  "message": "Device revoked successfully"
}
```

---

## Segurança (E2EE)

### Setup de Chaves (Upload)
Envia a chave mestra criptografada para o servidor (primeiro uso).

**Endpoint:** `POST /api/v1/security/keys/setup`

**Body:**
```json
{
  "encrypted_key": "base64_blob_encrypted_master_key...",
  "key_salt": "base64_salt_used_for_kdf...",
  "kdf_params": "{\"algorithm\":\"argon2id\",\"iterations\":3,\"memory\":65536,\"parallelism\":4}",
  "encryption_algo": "AES-256-GCM"
}
```

**Response (200 OK):**
```json
{
  "message": "Key uploaded successfully"
}
```

### Sync de Chaves (Download)
Baixa a chave mestra criptografada para descriptografia local.

**Endpoint:** `GET /api/v1/security/keys/sync`

**Response (200 OK):**
```json
{
  "encrypted_key": "base64_blob_encrypted_master_key...",
  "key_salt": "base64_salt_used_for_kdf...",
  "kdf_params": "{\"algorithm\":\"argon2id\",\"iterations\":3,\"memory\":65536,\"parallelism\":4}",
  "encryption_algo": "AES-256-GCM",
  "updated_at": "2023-11-28T12:00:00Z"
}
```

---

## Notas

### Criar Nota
Cria uma nova nota ou diretório. O conteúdo deve ser criptografado pelo cliente.

**Endpoint:** `POST /api/v1/notes`

**Body:**
```json
{
  "type": "file",
  "parent_id": null,
  "encrypted_title": "base64_encrypted_title...",
  "encrypted_content": "base64_encrypted_content...",
  "encryption_algo": "AES-256-GCM",
  "nonce": "base64_nonce..."
}
```

**Response (201 Created):**
```json
{
  "id": "note:uuid...",
  "parent_id": null,
  "type": "file",
  "encrypted_title": "base64_encrypted_title...",
  "encrypted_content": "base64_encrypted_content...",
  "encryption_algo": "AES-256-GCM",
  "nonce": "base64_nonce...",
  "created_at": "2023-11-28T12:00:00Z",
  "updated_at": "2023-11-28T12:00:00Z",
  "is_deleted": false,
  "version": 1
}
```

### Listar Notas
Retorna todas as notas do usuário.

**Endpoint:** `GET /api/v1/notes`

**Response (200 OK):**
```json
[
  {
    "id": "note:uuid...",
    "parent_id": null,
    "type": "file",
    "encrypted_title": "base64_encrypted_title...",
    "version": 1,
    "updated_at": "2023-11-28T12:00:00Z"
    // ... outros campos
  }
]
```

### Obter Nota
Retorna detalhes de uma nota específica.

**Endpoint:** `GET /api/v1/notes/{id}`

**Response (200 OK):**
```json
{
  "id": "note:uuid...",
  "type": "file",
  "encrypted_title": "base64_encrypted_title...",
  "encrypted_content": "base64_encrypted_content...",
  // ...
}
```

### Atualizar Nota
Atualiza o conteúdo ou metadados de uma nota.

**Endpoint:** `PUT /api/v1/notes/{id}`

**Body:**
```json
{
  "encrypted_content": "new_base64_content...",
  "nonce": "new_nonce..."
}
```

**Response (200 OK):**
```json
{
  "id": "note:uuid...",
  "version": 2,
  "updated_at": "2023-11-28T12:05:00Z",
  // ...
}
```

### Deletar Nota
Marca uma nota como deletada (soft delete).

**Endpoint:** `DELETE /api/v1/notes/{id}`

**Response (200 OK):**
```json
{
  "message": "Note deleted successfully"
}
```

---

## Sincronização Eficiente

### Get Manifest
Retorna uma lista compacta de todas as notas para comparação eficiente.
Este endpoint é otimizado para sync inicial e verificação de estado.

**Endpoint:** `GET /api/v1/sync/manifest`

**Response (200 OK):**
```json
{
  "notes": [
    {
      "id": "note:uuid...",
      "content_hash": "sha256_base64...",
      "version": 5,
      "updated_at": "2023-11-28T12:05:00Z",
      "is_deleted": false
    }
  ],
  "sync_time": "2023-11-28T12:10:00Z"
}
```

### Batch Diff
Compara o estado local do cliente com o servidor e retorna ações necessárias.
Permite sync eficiente em uma única requisição.

**Endpoint:** `POST /api/v1/sync/batch-diff`

**Body:**
```json
{
  "device_id": "device:uuid...",
  "local_notes": [
    {
      "id": "note:uuid...",
      "content_hash": "sha256_base64...",
      "version": 3
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "to_download": [
    {
      "id": "note:uuid...",
      "encrypted_title": "base64...",
      "encrypted_content": "base64...",
      "version": 5,
      // ... full note data
    }
  ],
  "to_upload": ["note:uuid1...", "note:uuid2..."],
  "to_delete": ["note:uuid3..."],
  "conflicts": [
    {
      "note_id": "note:uuid4...",
      "local_hash": "abc123...",
      "server_hash": "def456...",
      "local_version": 3,
      "server_version": 3
    }
  ],
  "sync_time": "2023-11-28T12:10:00Z"
}
