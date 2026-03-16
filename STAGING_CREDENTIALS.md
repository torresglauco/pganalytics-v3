# 🔐 Phase 4 Staging Environment - Credenciais de Acesso

## ✅ Credenciais de Login

### Frontend / API
**URL**: http://localhost:3000/login

**Credenciais**:
- **Email**: `admin@pganalytics.com`
- **Username**: `admin`
- **Senha**: `Admin@123456`

✅ **Testado e funcionando!**

---

## 📊 Grafana Dashboards
**URL**: http://localhost:3001

- **Username**: `admin`
- **Senha**: `grafana-1773664938`

---

## 🗄️ PostgreSQL Database
**Conexão**:
- **Host**: localhost
- **Port**: 5432
- **Database**: pganalytics_staging
- **Username**: pganalytics
- **Senha**: `staging-1773664938`

**Comando de Conexão**:
```bash
psql -h localhost -U pganalytics -d pganalytics_staging
# Password: staging-1773664938
```

---

## 🔗 Outros Acessos

### Redis
- **Host**: localhost
- **Port**: 6379
- Sem autenticação

### Prometheus
- **URL**: http://localhost:9090
- Acesso público

---

## 🧪 API REST

### Health Check
```bash
curl http://localhost:8000/api/v1/health
```

**Resposta esperada**:
```json
{
  "status": "ok",
  "version": "3.0.0-alpha",
  "database_ok": true,
  "timescale_ok": true
}
```

### Login (Obter JWT Token)
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "Admin@123456"
  }'
```

**Resposta**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@pganalytics.com",
    "role": "admin"
  }
}
```

### Usar Token para Acessar API Protegida
```bash
TOKEN="seu-token-aqui"
curl http://localhost:3000/api/v1/alerts \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📋 Resumo de Acesso

| Serviço | URL | Tipo | Credencial |
|---------|-----|------|-----------|
| **Frontend** | http://localhost:3000 | Web | admin / Admin@123456 |
| **API** | http://localhost:8000/api/v1/ | REST | JWT Token |
| **Grafana** | http://localhost:3001 | Web | admin / grafana-1773664938 |
| **Prometheus** | http://localhost:9090 | Web | Público |
| **PostgreSQL** | localhost:5432 | DB | pganalytics / staging-1773664938 |
| **Redis** | localhost:6379 | Cache | Sem auth |

---

## 🚀 Próximos Passos

1. **Acesse o Frontend**:
   - Abra http://localhost:3000/login
   - Use as credenciais acima
   - Teste os menus e funcionalidades do Phase 4

2. **Explore o Grafana**:
   - Acesse http://localhost:3001
   - Visualize as métricas do Prometheus

3. **Teste a API**:
   - Use o token JWT obtido para acessar endpoints protegidos
   - Teste os endpoints de Phase 4 (Alerts, Silencing, Escalation)

---

**Data**: 2026-03-16  
**Versão**: Phase 4 v4.0.0  
**Status**: ✅ Pronto para uso
