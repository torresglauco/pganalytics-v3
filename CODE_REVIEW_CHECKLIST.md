# Code Review Checklist

## Overview

This checklist helps reviewers validate the migration runner and password change implementation.

---

## Migration Runner Code Review

### File: `/backend/internal/storage/migrations.go`

#### Structure ✅
- [ ] Package declaration correct
- [ ] All imports present and necessary
- [ ] No circular imports
- [ ] MigrationRunner struct properly defined

#### Main Methods
- [ ] `NewMigrationRunner()` - Creates instance correctly
- [ ] `Run()` - Orchestrates full migration flow
- [ ] `createVersionsTable()` - Uses IF NOT EXISTS for idempotence
- [ ] `loadMigrations()` - Handles multiple path options
- [ ] `executeMigration()` - Uses transactions correctly
- [ ] `GetExecutedMigrations()` - Returns proper data structure

#### Error Handling
- [ ] All errors wrapped with context
- [ ] No silent failures
- [ ] Logging at appropriate levels (Info, Debug, Error)
- [ ] Graceful handling of missing migrations directory

#### Transaction Safety
- [ ] Each migration wrapped in transaction
- [ ] Rollback on error
- [ ] Commit only after successful execution
- [ ] schema_versions recorded in same transaction

#### SQL Statement Handling
- [ ] Supports multiple statements per file
- [ ] Handles quoted strings correctly
- [ ] Skips empty statements
- [ ] Provides helpful error messages with context

### Integration Points
- [ ] Called from `NewPostgresDB()` after connection
- [ ] Logger passed from main.go
- [ ] Doesn't block API startup on failure
- [ ] Works with existing connection pool

---

## Password Change Feature Review

### File: `/backend/migrations/018_password_changed.sql`

#### SQL Syntax
- [ ] ALTER TABLE syntax correct
- [ ] Column type is BOOLEAN with DEFAULT false
- [ ] IF NOT EXISTS prevents errors on re-run
- [ ] Index syntax valid
- [ ] Comments properly formatted

#### Safety
- [ ] Non-breaking change (new column)
- [ ] Has default value (handles existing rows)
- [ ] Index helps query performance
- [ ] Audit log entry included

### File: `/backend/pkg/models/models.go`

#### User Struct
- [ ] New PasswordChanged field added
- [ ] Proper type (bool)
- [ ] Correct db tag: `db:"password_changed"`
- [ ] JSON serialization: `json:"password_changed"`
- [ ] Field is exported (capitalized)
- [ ] Positioned logically with other security fields

### File: `/backend/internal/storage/postgres.go`

#### Query Updates
- [ ] GetUserByUsername includes password_changed
- [ ] GetUserByID includes password_changed  
- [ ] ListUsers includes password_changed
- [ ] CreateUserWithRole sets password_changed = false
- [ ] All Scan() calls match SELECT fields

#### New Method: UpdateUserPassword()
- [ ] Sets password_hash to new hash
- [ ] Sets password_changed = true
- [ ] Updates updated_at timestamp
- [ ] Proper error handling
- [ ] Returns error if user not found
- [ ] Uses context parameter

#### Existing Method: ResetUserPassword()
- [ ] NOT modified (admin use only)
- [ ] Still sets password but NOT password_changed flag

### File: `/backend/internal/api/handlers_auth.go`

#### New Handler: handleCheckPasswordChangeRequired()
- [ ] Gets user from context
- [ ] Fetches fresh data from database
- [ ] Checks password_changed field
- [ ] Returns proper JSON response
- [ ] Has proper HTTP status codes
- [ ] Error messages are user-friendly

#### Route Registration
- [ ] GET method correct
- [ ] Path: /api/v1/auth/password-change-required
- [ ] AuthMiddleware required
- [ ] Registered in RegisterAuthHandlers

#### Type Definition
- [ ] PasswordChangeRequiredResponse has proper JSON tags
- [ ] Fields are necessary and sufficient

### File: `/backend/internal/api/handlers.go`

#### Updated Method: handleChangePassword()
- [ ] Changed from ResetUserPassword to UpdateUserPassword
- [ ] Sets password_changed = true on successful change
- [ ] Existing error handling preserved
- [ ] Logging updated appropriately

---

## Integration Tests

### Database Tests
- [ ] schema_versions table can be created
- [ ] Password_changed column can be added
- [ ] Existing users get default value (false)
- [ ] New users created with password_changed = false
- [ ] UpdateUserPassword sets password_changed = true
- [ ] ResetUserPassword does NOT set password_changed

### API Tests
- [ ] New user registration sets password_changed = false
- [ ] password-change-required returns true for new user
- [ ] password-change-required returns false after change
- [ ] Password change endpoint works
- [ ] Old password verification works
- [ ] Proper error messages on failure

### Migration Tests
- [ ] Migrations directory found correctly
- [ ] Migration files loaded in order
- [ ] schema_versions table created
- [ ] Each migration recorded after execution
- [ ] Re-running doesn't execute twice
- [ ] Missing migrations directory doesn't crash

---

## Security Review

### Password Management
- [ ] Password hashed before storage
- [ ] Old password verified before allowing change
- [ ] New password validated (length, complexity if applicable)
- [ ] No password returned in API response
- [ ] password_changed field cannot be manually set

### Access Control
- [ ] password-change-required requires authentication
- [ ] Users can only check their own password status
- [ ] Password change enforced for new users
- [ ] Admin can still reset passwords without change

### Data Protection
- [ ] password_changed field not exposed unless authenticated
- [ ] Audit log entry created for schema changes
- [ ] Transaction isolation prevents race conditions
- [ ] Index prevents table scans for password change checks

---

## Performance Review

### Migration Performance
- [ ] Migrations cached after first execution
- [ ] schema_versions lookup is O(1)
- [ ] Parallel migration execution not attempted (sequential is fine)
- [ ] Large migration files handled gracefully

### Password Change Performance
- [ ] password_changed index created (WHERE password_changed = false)
- [ ] Query on password_changed uses index
- [ ] No N+1 query problems
- [ ] Connection pool not exhausted by migrations

---

## Backwards Compatibility

### Existing Code
- [ ] No breaking changes to User model (fields are additive)
- [ ] Existing password endpoints still work
- [ ] GetUserByID still works (null password_changed handled)
- [ ] No database migrations removed

### Graceful Degradation
- [ ] Migrations directory missing doesn't crash
- [ ] password_changed column missing doesn't crash queries
- [ ] Old databases work with new code

---

## Documentation Review

### Code Comments
- [ ] functions documented with godoc format
- [ ] Complex logic has inline comments
- [ ] Migration SQL includes explanatory comments

### User Documentation
- [ ] Password change flow documented
- [ ] Migration system documented
- [ ] Deployment instructions clear
- [ ] API endpoint documentation provided

---

## Testing Coverage

### Unit Tests Needed
- [ ] MigrationRunner.Run() success case
- [ ] MigrationRunner.Run() with missing directory
- [ ] MigrationRunner.Run() with invalid SQL
- [ ] MigrationRunner.Run() idempotence
- [ ] UpdateUserPassword() basic case
- [ ] UpdateUserPassword() user not found

### Integration Tests Needed
- [ ] Full password change flow (create user, change password)
- [ ] Multiple migrations execute in order
- [ ] password-change-required endpoint workflow
- [ ] Database state validation after migrations

---

## Deployment Checklist

### Pre-Deployment
- [ ] Code reviewed and approved
- [ ] Compilation verified
- [ ] All tests passing
- [ ] Database backup created
- [ ] Rollback plan documented

### Deployment
- [ ] Update code
- [ ] Mount migrations directory
- [ ] Set MIGRATIONS_PATH if needed
- [ ] Start new API version
- [ ] Monitor logs for migration errors
- [ ] Verify schema_versions table populated

### Post-Deployment
- [ ] Verify migrations executed
- [ ] Test password change flow
- [ ] Monitor error rates
- [ ] Check user complaints

### Frontend Deployment
- [ ] Implement password-change-required check
- [ ] Add password change modal
- [ ] Test with new user account
- [ ] Test with existing user account
- [ ] Verify user experience

---

## Sign-Off

- [ ] Code Review Approved: ___________
- [ ] Security Review Approved: ___________
- [ ] QA Approved: ___________
- [ ] DevOps Approved: ___________
- [ ] Ready for Production: YES / NO

---

## Notes & Comments

[Space for reviewer comments]

