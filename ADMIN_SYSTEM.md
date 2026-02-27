# pgAnalytics - Admin System Documentation

## Overview

pgAnalytics agora utiliza um sistema de login administrativo automático onde:

- **Usuário Padrão:** `admin`
- **Senha Padrão:** `admin`
- **Auto-login:** O frontend faz login automático com essas credenciais no primeiro carregamento
- **Controle de Acesso:** Apenas usuários com role `admin` podem criar novos usuários
- **Form de Cadastro Removido:** A página inicial NÃO possui mais um formulário de cadastro público

## Arquitetura

### Backend

#### Modelos Novos
- `CreateUserRequest` - Modelo para criar usuários (com parâmetro `role`)
- Nova função `CreateUserWithRole()` em PostgresDB

#### Endpoints Novos
- `POST /api/v1/users` - Criar novo usuário (requer autenticação, apenas admin)

#### Modificações
- Middleware AuthMiddleware agora popula o objeto `User` completo no contexto (não apenas IDs)
- Signup removido da lista de rotas públicas (routes ainda existe, mas não é exposto)

### Frontend

#### Modificações em App.tsx
- Auto-login automático com `admin:admin`
- Mostra tela de carregamento durante auto-login
- Se falhar, mostra erro com opção de retry

#### Novo Componente CreateUserForm
- Localizado em `src/components/CreateUserForm.tsx`
- Formulário para criar novos usuários
- Suporta dois tipos de perfil: `user` (Regular User) e `admin` (Administrator)
- Validação em tempo real
- Integração com API `/api/v1/users`

#### Modificações em Dashboard.tsx
- Nova aba "Create User" (visível apenas para admins)
- Admin pode criar usuários normais ou outros admins
- Mensagens de sucesso/erro ao criar usuários
- Detecção automática de perfil (`currentUser?.role === 'admin'`)

## Fluxo de Uso

### Primeiro Login
1. Abrir http://localhost:4000
2. Frontend faz auto-login com `admin:admin`
3. Usuário é redirecionado para Dashboard
4. Admin vê o painel completo incluindo aba "Create User"

### Criar Novo Usuário (como Admin)
1. Na aba "Create User" do Dashboard
2. Preencher:
   - Username (3+ caracteres)
   - Email (formato válido)
   - Password (8+ caracteres)
   - Full Name (opcional)
   - Role (user ou admin)
3. Clicar "Create User"
4. Mensagem de sucesso aparece

### Usuários Criados
- Podem fazer login via POST /api/v1/auth/login
- Se role="user": acesso limitado (sem aba "Create User")
- Se role="admin": acesso total (com aba "Create User")

## Status de Implementação

### ✅ Implementação Completa

A implementação foi finalizada e testada com sucesso. Todos os componentes estão funcionando:

1. **Backend**:
   - ✅ Auto-login com credenciais admin:admin funcionando
   - ✅ Criação de usuários via POST /api/v1/users
   - ✅ Controle de acesso baseado em role (admin-only)
   - ✅ Suporte para dois tipos de perfil (admin e user)

2. **Frontend**:
   - ✅ Proxy Node.js integrando frontend ao backend
   - ✅ Auto-login implementado
   - ✅ Formulário de criação de usuários (admin-only)
   - ✅ Interface responsiva e validação de formulário

3. **Docker**:
   - ✅ Frontend, Backend e Databases em containers
   - ✅ Comunicação entre serviços via Docker network
   - ✅ Health checks configurados
   - ✅ Portas mapeadas corretamente

### Problema Resolvido
- **Issue**: POST /api/v1/users retornava 401 mesmo com token válido
- **Causa**: Middleware inline não estava sendo executado antes do handler
- **Solução**: Aplicar middleware via `Group.Use()` ao invés de inline
- **Resultado**: Todos os testes passando (6/6)

## Configuração do Banco de Dados

O usuário admin é criado automaticamente pela migração `001_init.sql`:

```sql
INSERT INTO users (username, email, password_hash, full_name, role)
VALUES (
    'admin',
    'admin@pganalytics.local',
    crypt('admin', gen_salt('bf')),
    'Administrator',
    'admin'
) ON CONFLICT DO NOTHING;
```

## Segurança

- Senhas são hashadas com bcrypt (cost factor 10)
- Apenas admins podem criar usuários
- Tokens JWT com expiração de 15 minutos (access) e 7 dias (refresh)
- Validação em servidor (não confiar apenas em validação cliente)

## Rollback se Necessário

Se desejar voltar ao sistema de signup público:
1. Restaurar `auth.POST("/signup", s.handleSignup)` em server.go
2. Remover `users` group das rotas
3. Remover CreateUserForm do Dashboard
4. Restaurar App.tsx para mostrar AuthPage ao invés de auto-login

## Testes e Validação

Todos os testes abaixo foram executados e passaram com sucesso:

### ✅ Testes Realizados
1. **Frontend Acessível**: http://localhost:4000
2. **Admin Login**: ✓ Sucesso com admin:admin
3. **Regular User Login**: ✓ Sucesso (john_doe criado durante testes)
4. **Admin User Creation**: ✓ Admins podem criar novos usuários
5. **Role-Based Access Control**: ✓ Usuários normais recebem 403 ao tentar criar usuários
6. **Admin User Creation**: ✓ Admins podem criar outros admins

### Acessar o Sistema

1. **Frontend UI**: http://localhost:4000
2. **Backend API**: http://localhost:8080/api/v1
3. **Grafana Dashboard**: http://localhost:3000 (admin/Th101327!!!)
4. **Usuário Padrão**: admin/admin

### Exemplo de Uso via curl

```bash
# Login como admin
TOKEN=$(curl -s -X POST http://localhost:4000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Criar novo usuário
curl -X POST http://localhost:4000/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"username":"newuser","email":"new@example.com","password":"SecurePass123","full_name":"New User","role":"user"}'
```

## Próximas Features (Future)

- [ ] Senha padrão obrigatória no primeiro login
- [ ] Listagem de usuários para admins
- [ ] Edição de usuários
- [ ] Deletar usuários
- [ ] Resetar senhas
- [ ] Auditoria de criação de usuários
