# Migration Runner & Password Change - Quick Start Guide

## What Changed

### 1. Automatic Database Migrations
Your database migrations in `/backend/migrations/` are now **automatically executed** when the API starts.

### 2. Mandatory First-Login Password Change
New users must **change their password on first login** for security.

## For Backend Developers

### Running Migrations
No action needed! Migrations run automatically on startup:

```go
// In NewPostgresDB()
if err := runMigrations(ctx, db); err != nil {
    fmt.Fprintf(os.Stderr, "Warning: Migration execution error: %v\n", err)
}
```

### Creating a New Migration
1. Create file in `/backend/migrations/` with format: `NNN_description.sql`
2. File is executed automatically on next startup

Example:
```sql
-- /backend/migrations/019_my_feature.sql
CREATE TABLE my_table (
    id SERIAL PRIMARY KEY,
    ...
);
```

### Checking Migration History
```bash
psql $DATABASE_URL -c "SELECT * FROM pganalytics.schema_versions;"
```

---

## For Frontend Developers

### Password Change Flow (NEW)

After user login, check if password change is required:

```javascript
// Step 1: After login, check requirement
const response = await fetch('/api/v1/auth/password-change-required', {
  headers: { 'Authorization': `Bearer ${accessToken}` }
});

const data = await response.json();
if (data.password_change_required) {
  // Show password change modal
}
```

API Response:
```json
{
  "password_change_required": true,
  "message": "Password change is required on first login"
}
```

### Password Change Endpoint

```
POST /api/v1/auth/change-password
Authorization: Bearer <token>

{
  "old_password": "current_password",
  "new_password": "new_secure_password"
}
```

Response:
```json
{
  "message": "Password changed successfully"
}
```

### Suggested Modal Implementation

```jsx
function PasswordChangeModal({ open, onSuccess, accessToken }) {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');

  const handleChangePassword = async () => {
    if (newPassword !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    try {
      const response = await fetch('/api/v1/auth/change-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken}`
        },
        body: JSON.stringify({
          old_password: currentPassword,
          new_password: newPassword
        })
      });

      if (response.ok) {
        onSuccess();
      } else {
        const error = await response.json();
        setError(error.message || 'Failed to change password');
      }
    } catch (e) {
      setError('Network error: ' + e.message);
    }
  };

  return (
    <Modal open={open} title="Set Your Password">
      <Input
        type="password"
        label="Current Password"
        value={currentPassword}
        onChange={(e) => setCurrentPassword(e.target.value)}
      />
      <Input
        type="password"
        label="New Password"
        value={newPassword}
        onChange={(e) => setNewPassword(e.target.value)}
      />
      <Input
        type="password"
        label="Confirm Password"
        value={confirmPassword}
        onChange={(e) => setConfirmPassword(e.target.value)}
      />
      {error && <Alert type="error">{error}</Alert>}
      <Button onClick={handleChangePassword}>Change Password</Button>
    </Modal>
  );
}
```

---

## For DevOps/Docker

### Mounting Migrations Directory

In your Dockerfile:
```dockerfile
COPY ./backend/migrations /app/migrations
ENV MIGRATIONS_PATH=/app/migrations
```

Or in docker-compose.yml:
```yaml
services:
  api:
    volumes:
      - ./backend/migrations:/app/migrations:ro
    environment:
      MIGRATIONS_PATH: /app/migrations
```

### Checking Migrations in Running Container

```bash
docker exec <container> psql $DATABASE_URL \
  -c "SELECT version, executed_at FROM pganalytics.schema_versions;"
```

---

## For QA/Testing

### Test Migration Execution
```bash
# 1. Connect to database
psql $DATABASE_URL

# 2. Check schema_versions table
SELECT * FROM pganalytics.schema_versions ORDER BY executed_at DESC;

# 3. Verify tables exist
SELECT table_name FROM information_schema.tables
  WHERE table_schema='pganalytics' ORDER BY table_name;
```

### Test Password Change Flow
```bash
# 1. Setup new user
curl -X POST http://localhost:8080/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "InitialPass123",
    "full_name": "Test User"
  }'
# Save the access_token from response

# 2. Check password change requirement (should be true)
curl http://localhost:8080/api/v1/auth/password-change-required \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

# 3. Change password
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -d '{
    "old_password": "InitialPass123",
    "new_password": "NewPass123!@"
  }'

# 4. Verify requirement is now false
curl http://localhost:8080/api/v1/auth/password-change-required \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

---

## Common Issues & Solutions

### Issue: Migrations not running
**Solution**: Check `MIGRATIONS_PATH` environment variable points to correct directory
```bash
echo $MIGRATIONS_PATH
ls -la $MIGRATIONS_PATH
```

### Issue: "Table doesn't exist" error
**Solution**: Verify migrations executed successfully
```sql
SELECT COUNT(*) FROM pganalytics.schema_versions;
```

### Issue: User created with old password scheme
**Solution**: Password change still works - just uses UpdateUserPassword instead of ResetUserPassword

### Issue: Password change returns "current password is incorrect"
**Solution**: User must provide the actual current password, not a new one

---

## Key Files to Review

| File | Purpose |
|------|---------|
| `/backend/internal/storage/migrations.go` | Migration runner (NEW) |
| `/backend/migrations/018_password_changed.sql` | Password change feature (NEW) |
| `/backend/internal/storage/postgres.go` | Updated with UpdateUserPassword() |
| `/backend/pkg/models/models.go` | Added PasswordChanged field |
| `/backend/internal/api/handlers_auth.go` | Added password-change-required endpoint |

---

## Deployment Checklist

- [ ] Update Docker/container to mount migrations directory
- [ ] Verify MIGRATIONS_PATH environment variable is set
- [ ] Run migrations on first startup (automatic)
- [ ] Check schema_versions table is populated
- [ ] Frontend implements password change modal
- [ ] Frontend calls password-change-required endpoint after login
- [ ] Test with new user account (setup endpoint)
- [ ] Test password change flow end-to-end
- [ ] Verify old password verification works
- [ ] Check error messages are user-friendly

