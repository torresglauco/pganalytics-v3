package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// TENANT MODELS (SCALE-01, SCALE-02, SCALE-03, SCALE-04)
// ============================================================================

// Tenant represents a tenant in the multi-tenant SaaS architecture
// Each tenant has isolated access to their own collectors and data via RLS
type Tenant struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"` // URL-friendly identifier
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// TenantUser represents the relationship between a user and a tenant
// A user can belong to multiple tenants with different roles
type TenantUser struct {
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"` // admin, viewer, editor
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TenantCollectorMapping represents the assignment of a collector to a tenant
// Collectors are scoped to a single tenant for data isolation
type TenantCollectorMapping struct {
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	CollectorID uuid.UUID `json:"collector_id" db:"collector_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TenantResponse represents the API response for tenant data
type TenantResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantListResponse contains a list of tenants with metadata
type TenantListResponse struct {
	Count   int               `json:"count"`
	Tenants []*TenantResponse `json:"tenants"`
}

// TenantCreateRequest represents the request body for creating a new tenant
type TenantCreateRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

// TenantCollectorAssignmentRequest represents the request body for assigning a collector to a tenant
type TenantCollectorAssignmentRequest struct {
	CollectorID uuid.UUID `json:"collector_id" binding:"required"`
}
