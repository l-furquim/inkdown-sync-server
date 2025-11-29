#!/bin/bash

# Script de teste da API de autenticação
BASE_URL="http://localhost:8080"

echo "=== Inkdown Sync Server - Testes de Autenticação ==="
echo ""

# 1. Teste de Registro
echo "1. Testando registro de usuário..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@inkdown.com",
    "password": "Test123!@#"
  }')

echo "Response: $REGISTER_RESPONSE"
echo ""

# 2. Teste de Login
echo "2. Testando login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \  -H "Content-Type: application/json" \
  -d '{
    "email": "test@inkdown.com",
    "password": "Test123!@#"
  }')

echo "Response: $LOGIN_RESPONSE"
echo ""

# Extrair token (requer jq)
if command -v jq &> /dev/null; then
    ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.access_token')
    echo "Access Token: $ACCESS_TOKEN"
else
    echo "jq não instalado, extrair token manualmente"
    ACCESS_TOKEN="REPLACE_WITH_TOKEN"
fi
echo ""

# 3. Teste de rota protegida - GET /users/me
echo "3. Testando rota protegida (GET /users/me)..."
curl -s -X GET "$BASE_URL/api/v1/users/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
echo ""
echo ""

# 4. Teste sem token (deve falhar)
echo "4. Testando acesso sem token (deve retornar 401)..."
curl -s -X GET "$BASE_URL/api/v1/users/me"
echo ""
echo ""

# 5. Teste de atualização de perfil
echo "5. Testando atualização de username..."
curl -s -X PUT "$BASE_URL/api/v1/users/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"username": "updateduser"}'
echo ""
echo ""

echo "=== Testes concluídos ==="
