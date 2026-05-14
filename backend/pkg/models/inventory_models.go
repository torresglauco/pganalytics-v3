package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// INVENTORY MODELS
// ============================================================================

// TableInventory represents database table inventory with size and usage metrics
// INV-01: User can view table inventory with row counts and sizes
type TableInventory struct {
	Time         time.Time `json:"time" db:"time"`
	CollectorID  uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName string    `json:"database_name" db:"database_name"`
	SchemaName   string    `json:"schema_name" db:"schema_name"`
	TableName    string    `json:"table_name" db:"table_name"`
	TableType    string    `json:"table_type" db:"table_type"` // BASE TABLE, VIEW
	RowCount     int64     `json:"row_count" db:"row_count"`   // from n_live_tup
	TotalSizeMb  int64     `json:"total_size_mb" db:"total_size_mb"`
	TableSizeMb  int64     `json:"table_size_mb" db:"table_size_mb"`
	IndexSizeMb  int64     `json:"index_size_mb" db:"index_size_mb"`
	ToastSizeMb  int64     `json:"toast_size_mb" db:"toast_size_mb"`
	HasOids      bool      `json:"has_oids" db:"has_oids"`
	TableOid     uint64    `json:"table_oid" db:"table_oid"` // for change detection
}

// ColumnInventory represents database column inventory with type information
// INV-02: User can view column inventory with data types and nullability
type ColumnInventory struct {
	Time               time.Time `json:"time" db:"time"`
	CollectorID        uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName       string    `json:"database_name" db:"database_name"`
	SchemaName         string    `json:"schema_name" db:"schema_name"`
	TableName          string    `json:"table_name" db:"table_name"`
	ColumnName         string    `json:"column_name" db:"column_name"`
	DataType           string    `json:"data_type" db:"data_type"`
	IsNullable         bool      `json:"is_nullable" db:"is_nullable"`
	ColumnDefault      string    `json:"column_default" db:"column_default"`
	OrdinalPosition    int       `json:"ordinal_position" db:"ordinal_position"`
	CharacterMaxLength int       `json:"character_max_length" db:"character_max_length"`
	NumericPrecision   int       `json:"numeric_precision" db:"numeric_precision"`
	NumericScale       int       `json:"numeric_scale" db:"numeric_scale"`
	IsPrimaryKey       bool      `json:"is_primary_key" db:"is_primary_key"`
	IsForeignKey       bool      `json:"is_foreign_key" db:"is_foreign_key"`
}

// IndexInventory represents database index inventory with usage statistics
// INV-03: User can view index inventory with usage statistics
type IndexInventory struct {
	Time            time.Time `json:"time" db:"time"`
	CollectorID     uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName    string    `json:"database_name" db:"database_name"`
	SchemaName      string    `json:"schema_name" db:"schema_name"`
	TableName       string    `json:"table_name" db:"table_name"`
	IndexName       string    `json:"index_name" db:"index_name"`
	IndexDefinition string    `json:"index_definition" db:"index_definition"`
	IndexSizeMb     int64     `json:"index_size_mb" db:"index_size_mb"`
	IdxScan         int64     `json:"idx_scan" db:"idx_scan"`
	IdxTupRead      int64     `json:"idx_tup_read" db:"idx_tup_read"`
	IdxTupFetch     int64     `json:"idx_tup_fetch" db:"idx_tup_fetch"`
	UsageStatus     string    `json:"usage_status" db:"usage_status"` // UNUSED, RARELY_USED, ACTIVE
	IsPrimary       bool      `json:"is_primary" db:"is_primary"`
	IsUnique        bool      `json:"is_unique" db:"is_unique"`
	IndexOid        uint64    `json:"index_oid" db:"index_oid"`
}

// ExtensionInventory represents database extension inventory
// INV-04: User can view extension inventory with versions
type ExtensionInventory struct {
	Time             time.Time `json:"time" db:"time"`
	CollectorID      uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName     string    `json:"database_name" db:"database_name"`
	ExtensionName    string    `json:"extension_name" db:"extension_name"`
	ExtensionVersion string    `json:"extension_version" db:"extension_version"`
	ExtensionOwner   string    `json:"extension_owner" db:"extension_owner"`
	ExtensionSchema  string    `json:"extension_schema" db:"extension_schema"`
	IsRelocatable    bool      `json:"is_relocatable" db:"is_relocatable"`
	Description      string    `json:"description" db:"description"`
}

// SchemaVersion represents schema change tracking
// INV-05: User can track schema changes over time
type SchemaVersion struct {
	Time          time.Time `json:"time" db:"time"`
	CollectorID   uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName  string    `json:"database_name" db:"database_name"`
	VersionHash   string    `json:"version_hash" db:"version_hash"` // MD5 of all object signatures
	ChangeType    string    `json:"change_type" db:"change_type"`   // INIT, TABLE_ADDED, TABLE_REMOVED, TABLE_MODIFIED, etc.
	ObjectType    string    `json:"object_type" db:"object_type"`   // TABLE, COLUMN, INDEX, EXTENSION
	ObjectName    string    `json:"object_name" db:"object_name"`
	ChangeDetails string    `json:"change_details" db:"change_details"` // JSON describing the change
	PreviousValue string    `json:"previous_value" db:"previous_value"`
	NewValue      string    `json:"new_value" db:"new_value"`
}

// InventoryMetricsResponse contains all inventory-related metrics
type InventoryMetricsResponse struct {
	TableInventory     []*TableInventory     `json:"table_inventory,omitempty"`
	ColumnInventory    []*ColumnInventory    `json:"column_inventory,omitempty"`
	IndexInventory     []*IndexInventory     `json:"index_inventory,omitempty"`
	ExtensionInventory []*ExtensionInventory `json:"extension_inventory,omitempty"`
	SchemaVersions     []*SchemaVersion      `json:"schema_versions,omitempty"`
}
