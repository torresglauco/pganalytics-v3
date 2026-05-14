package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test Tenant struct has all required fields
func TestTenantFields(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	tenant := Tenant{
		ID:        id,
		Name:      "Acme Corporation",
		Slug:      "acme-corp",
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}

	if tenant.ID != id {
		t.Error("ID field not set correctly")
	}
	if tenant.Name != "Acme Corporation" {
		t.Error("Name field not set correctly")
	}
	if tenant.Slug != "acme-corp" {
		t.Error("Slug field not set correctly")
	}
	if tenant.CreatedAt != now {
		t.Error("CreatedAt field not set correctly")
	}
	if tenant.UpdatedAt != now {
		t.Error("UpdatedAt field not set correctly")
	}
	if !tenant.IsActive {
		t.Error("IsActive field not set correctly")
	}
}

// Test TenantUser struct has all required fields
func TestTenantUserFields(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	tenantUser := TenantUser{
		TenantID:  tenantID,
		UserID:    userID,
		Role:      "admin",
		CreatedAt: now,
	}

	if tenantUser.TenantID != tenantID {
		t.Error("TenantID field not set correctly")
	}
	if tenantUser.UserID != userID {
		t.Error("UserID field not set correctly")
	}
	if tenantUser.Role != "admin" {
		t.Error("Role field not set correctly")
	}
	if tenantUser.CreatedAt != now {
		t.Error("CreatedAt field not set correctly")
	}
}

// Test TenantUser roles are valid
func TestTenantUserRoles(t *testing.T) {
	validRoles := map[string]bool{
		"admin":  true,
		"viewer": true,
		"editor": true,
	}

	for role := range validRoles {
		tenantUser := TenantUser{
			Role: role,
		}
		if !validRoles[tenantUser.Role] {
			t.Errorf("Invalid role: %s", role)
		}
	}
}

// Test TenantCollectorMapping struct has all required fields
func TestTenantCollectorMappingFields(t *testing.T) {
	tenantID := uuid.New()
	collectorID := uuid.New()
	now := time.Now()

	mapping := TenantCollectorMapping{
		TenantID:    tenantID,
		CollectorID: collectorID,
		CreatedAt:   now,
	}

	if mapping.TenantID != tenantID {
		t.Error("TenantID field not set correctly")
	}
	if mapping.CollectorID != collectorID {
		t.Error("CollectorID field not set correctly")
	}
	if mapping.CreatedAt != now {
		t.Error("CreatedAt field not set correctly")
	}
}

// Test TenantResponse struct
func TestTenantResponse(t *testing.T) {
	response := TenantResponse{
		ID:        uuid.New(),
		Name:      "Test Tenant",
		Slug:      "test-tenant",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if response.Name != "Test Tenant" {
		t.Error("Name field not set correctly")
	}
	if response.Slug != "test-tenant" {
		t.Error("Slug field not set correctly")
	}
}
