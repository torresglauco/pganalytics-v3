# Teste Rápido - Delete Collector Fix

## ⚡ Como Testar a Correção

### Passo 1: Reconstruir a Aplicação (2 minutos)

```bash
# Remover containers antigos
docker-compose down -v

# Reconstruir e iniciar
docker-compose up -d --build

# Aguardar ~30 segundos para o backend estar pronto
sleep 30
```

### Passo 2: Executar Demo Setup (2 minutos)

```bash
./demo-setup.sh
```

Isso vai criar:
- ✅ Usuário demo (demo/Demo@12345)
- ✅ Um collector registrado
- ✅ Uma managed instance

### Passo 3: Iniciar Frontend (1 minuto)

```bash
./start-frontend.sh
```

### Passo 4: Testar no Navegador

1. Abra: **http://localhost:3000**
2. Login com:
   - Username: `demo`
   - Password: `Demo@12345`
3. Vá para aba: **"Active Collectors"**
4. Clique no botão 🗑️ **delete** do collector

### Resultado Esperado ✅

- O collector deve desaparecer da lista
- **NENHUMA** mensagem de erro
- A lista se atualiza automaticamente

---

## 📋 Checklist de Testes

- [ ] Backend compila sem erros
- [ ] Demo setup cria collector com sucesso
- [ ] Frontend carrega corretamente
- [ ] Login funciona com demo user
- [ ] Lista de collectors apareça
- [ ] Botão delete está visível
- [ ] Clicar delete remove o collector
- [ ] Sem erro "Not implemented yet"
- [ ] Sem erro "Error loading collectors"

---

## 🔧 Troubleshooting Rápido

Se algo não funcionar:

### Backend não inicia?
```bash
docker-compose logs backend
```

### Frontend não conecta?
```bash
# Verificar se backend está respondendo
curl http://localhost:8080/health

# Se falhar, reiniciar backend
docker-compose restart backend
```

### Collector não foi criado?
```bash
# Verificar logs do demo setup
docker-compose logs postgres
```

---

## ✅ Commits Feitos

```
d8f88f2 - feat: Implement GetCollector endpoint
b874094 - feat: Implement DeleteCollector endpoint
```

Os seguintes arquivos foram modificados:
- `backend/internal/api/handlers.go` - Implementação do handler
- `backend/internal/storage/postgres.go` - Método de delete no DB
- `backend/internal/storage/collector_store.go` - Wrapper do método

---

**Status**: ✅ Pronto para Testar

Para verificar o status das mudanças:
```bash
git log --oneline -5
git diff HEAD~3 HEAD
```
