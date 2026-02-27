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

## Implementação em Progresso

### Problemas Conhecidos
1. **Autenticação na Rota /api/v1/users**:
   - O endpoint está retornando 401 mesmo com token válido
   - Middleware parece não estar populando o contexto corretamente
   - TODO: Investigar Gin middleware ordering ou conflito com outros middlewares

2. **Próximos Passos**:
   - Verificar se há conflito com CollectorAuthMiddleware ou outros middlewares
   - Testar se rota /api/v1/users realmente existe e está registrada
   - Considerar usar `c.Set("user", user)` de forma diferente
   - Verificar se há cache de rotas no Gin que precisa ser limpo

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

## Próximas Features (Future)

- [ ] Senha padrão obrigatória no primeiro login
- [ ] Listagem de usuários para admins
- [ ] Edição de usuários
- [ ] Deletar usuários
- [ ] Resetar senhas
- [ ] Auditoria de criação de usuários
