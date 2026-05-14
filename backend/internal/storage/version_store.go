package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// VERSION DETECTION AND CAPABILITY TRACKING
// ============================================================================

// GetPostgreSQLVersion retrieves the PostgreSQL version for a collector
func (p *PostgresDB) GetPostgreSQLVersion(ctx context.Context, collectorID uuid.UUID) (*models.PostgreSQLVersion, error) {
	query := `
		SELECT postgres_version
		FROM collectors
		WHERE id = $1
	`

	var versionStr string
	err := p.db.QueryRowContext(ctx, query, collectorID).Scan(&versionStr)
	if err != nil {
		return nil, apperrors.DatabaseError("query collector version", err.Error())
	}

	// Parse version string (e.g., "16.2" or "16")
	major, minor, err := ParseVersion(versionStr)
	if err != nil {
		return nil, apperrors.BadRequest("invalid version format", err.Error())
	}

	// Build version info
	version := &models.PostgreSQLVersion{
		Major:       major,
		Minor:       minor,
		FullVersion: versionStr,
		IsSupported: IsVersionSupported(major),
	}

	// Set release and EOL dates based on major version
	version.ReleaseDate, version.EOLDate = getVersionDates(major)

	return version, nil
}

// GetVersionCapabilities returns the capabilities for a given PostgreSQL major version
func (p *PostgresDB) GetVersionCapabilities(ctx context.Context, version int) (*models.VersionCapabilities, error) {
	caps := &models.VersionCapabilities{
		Version: models.PostgreSQLVersion{
			Major:       version,
			IsSupported: IsVersionSupported(version),
		},
		// Version-specific capabilities
		HasWriteLagColumns:    version >= 13,
		HasWalReceiver:        version >= 10, // Actually available from 9.6
		HasLogicalReplication: version >= 10,
		HasPublication:        version >= 10,
		HasStandbySignal:      version >= 12,
		HasPgStatWal:          version >= 14, // Actually 14+ has pg_stat_wal
		HasPgStatSubscription: version >= 10,
	}

	// Set minimum query version
	if version >= 13 {
		caps.MinQueryVersion = "13.0"
	} else if version >= 10 {
		caps.MinQueryVersion = "10.0"
	} else {
		caps.MinQueryVersion = "9.4"
	}

	// Set release and EOL dates
	caps.Version.ReleaseDate, caps.Version.EOLDate = getVersionDates(version)
	caps.Version.FullVersion = fmt.Sprintf("%d", version)

	return caps, nil
}

// GetAllSupportedVersions returns all supported PostgreSQL versions
func (p *PostgresDB) GetAllSupportedVersions() []*models.PostgreSQLVersion {
	// PostgreSQL version lifecycle data
	// Source: https://www.postgresql.org/support/versioning/
	versions := []*models.PostgreSQLVersion{
		{Major: 17, Minor: 0, FullVersion: "17.0", IsSupported: true, ReleaseDate: parseDate("2024-09-26"), EOLDate: parseDate("2029-11-08")},
		{Major: 16, Minor: 0, FullVersion: "16.0", IsSupported: true, ReleaseDate: parseDate("2023-09-14"), EOLDate: parseDate("2028-11-09")},
		{Major: 15, Minor: 0, FullVersion: "15.0", IsSupported: true, ReleaseDate: parseDate("2022-10-13"), EOLDate: parseDate("2027-11-11")},
		{Major: 14, Minor: 0, FullVersion: "14.0", IsSupported: true, ReleaseDate: parseDate("2021-09-30"), EOLDate: parseDate("2026-11-12")},
		{Major: 13, Minor: 0, FullVersion: "13.0", IsSupported: true, ReleaseDate: parseDate("2020-09-24"), EOLDate: parseDate("2025-11-13")},
		// PostgreSQL 11 and 12 are EOL but we support detecting them
		{Major: 12, Minor: 0, FullVersion: "12.0", IsSupported: false, ReleaseDate: parseDate("2019-10-03"), EOLDate: parseDate("2024-11-14")},
		{Major: 11, Minor: 0, FullVersion: "11.0", IsSupported: false, ReleaseDate: parseDate("2018-10-18"), EOLDate: parseDate("2023-11-09")},
	}

	return versions
}

// IsVersionSupported checks if a PostgreSQL major version is currently supported
func IsVersionSupported(majorVersion int) bool {
	// PostgreSQL 13-17 are currently supported (as of 2024)
	// 11 and 12 are EOL (End of Life)
	return majorVersion >= 13 && majorVersion <= 17
}

// StoreCollectorVersion updates the postgres_version column for a collector
func (p *PostgresDB) StoreCollectorVersion(ctx context.Context, collectorID uuid.UUID, version string) error {
	query := `
		UPDATE collectors
		SET postgres_version = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := p.db.ExecContext(ctx, query, version, collectorID)
	if err != nil {
		return apperrors.DatabaseError("update collector version", err.Error())
	}

	return nil
}

// GetCollectorMode retrieves the deployment mode configuration for a collector
func (p *PostgresDB) GetCollectorMode(ctx context.Context, collectorID uuid.UUID) (*models.CollectorModeConfig, error) {
	query := `
		SELECT id, hostname, postgres_host, postgres_port, use_tls, tls_cert_file, tls_key_file, tls_ca_file
		FROM collectors
		WHERE id = $1
	`

	var (
		id           uuid.UUID
		hostname     string
		postgresHost string
		postgresPort int
		useTLS       bool
		tlsCertFile  sql.NullString
		tlsKeyFile   sql.NullString
		tlsCAFile    sql.NullString
	)

	err := p.db.QueryRowContext(ctx, query, collectorID).Scan(
		&id, &hostname, &postgresHost, &postgresPort, &useTLS,
		&tlsCertFile, &tlsKeyFile, &tlsCAFile,
	)
	if err != nil {
		return nil, apperrors.DatabaseError("query collector mode", err.Error())
	}

	// Determine mode based on postgres_host
	mode := "centralized"
	connectionType := "tcp"

	// Localhost or Unix socket connections indicate decentralized mode
	if postgresHost == "localhost" || postgresHost == "127.0.0.1" || strings.HasPrefix(postgresHost, "/") {
		mode = "decentralized"
		if strings.HasPrefix(postgresHost, "/") {
			connectionType = "unix_socket"
		}
	}

	config := &models.CollectorModeConfig{
		CollectorID:    id,
		Mode:           mode,
		ConnectionType: connectionType,
		UseTLS:         useTLS,
		TLSConfig: models.TLSConfig{
			CertFile: tlsCertFile.String,
			KeyFile:  tlsKeyFile.String,
			CAFile:   tlsCAFile.String,
		},
	}

	return config, nil
}

// ============================================================================
// VERSION HELPER FUNCTIONS
// ============================================================================

// ParseVersion parses a version string into major and minor components
// Examples: "16.2" -> (16, 2, nil), "16" -> (16, 0, nil)
func ParseVersion(versionStr string) (major, minor int, err error) {
	// Remove any trailing content after patch version
	parts := strings.Split(versionStr, ".")
	if len(parts) < 1 {
		return 0, 0, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	if len(parts) >= 2 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			// Minor version parsing failed, default to 0
			minor = 0
		}
	}

	return major, minor, nil
}

// CompareVersions compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	major1, minor1, _ := ParseVersion(v1)
	major2, minor2, _ := ParseVersion(v2)

	if major1 < major2 {
		return -1
	} else if major1 > major2 {
		return 1
	}

	// Major versions are equal, compare minor
	if minor1 < minor2 {
		return -1
	} else if minor1 > minor2 {
		return 1
	}

	return 0
}

// getVersionDates returns the release and EOL dates for a PostgreSQL major version
func getVersionDates(majorVersion int) (releaseDate, eolDate time.Time) {
	versions := map[int]struct {
		release string
		eol     string
	}{
		17: {"2024-09-26", "2029-11-08"},
		16: {"2023-09-14", "2028-11-09"},
		15: {"2022-10-13", "2027-11-11"},
		14: {"2021-09-30", "2026-11-12"},
		13: {"2020-09-24", "2025-11-13"},
		12: {"2019-10-03", "2024-11-14"},
		11: {"2018-10-18", "2023-11-09"},
	}

	if dates, ok := versions[majorVersion]; ok {
		return parseDate(dates.release), parseDate(dates.eol)
	}

	return time.Time{}, time.Time{}
}

// parseDate parses a date string in YYYY-MM-DD format
func parseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
