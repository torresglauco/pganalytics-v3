package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
)

// ============================================================================
// VERSION DETECTION TESTS
// ============================================================================

func TestPostgreSQLVersionDetection(t *testing.T) {
	testCases := []struct {
		name          string
		versionStr    string
		expectedMajor int
		expectedMinor int
	}{
		{"PostgreSQL 17", "17.0", 17, 0},
		{"PostgreSQL 16.2", "16.2", 16, 2},
		{"PostgreSQL 15.4", "15.4", 15, 4},
		{"PostgreSQL 14.1", "14.1", 14, 1},
		{"PostgreSQL 13.0", "13.0", 13, 0},
		{"PostgreSQL 12.18", "12.18", 12, 18},
		{"PostgreSQL 11.22", "11.22", 11, 22},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			major, minor, err := storage.ParseVersion(tc.versionStr)
			require.NoError(t, err, "ParseVersion should not error for %s", tc.versionStr)
			assert.Equal(t, tc.expectedMajor, major, "Major version mismatch")
			assert.Equal(t, tc.expectedMinor, minor, "Minor version mismatch")
		})
	}
}

func TestPostgreSQLVersionParsingEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		versionStr  string
		shouldError bool
	}{
		{"Empty string", "", true},
		{"Major only", "16", false},
		{"Patch version", "16.2.1", false},
		{"Invalid version", "invalid", true},
		{"Version with text", "16.2 (Debian)", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			major, minor, err := storage.ParseVersion(tc.versionStr)
			if tc.shouldError {
				assert.Error(t, err, "ParseVersion should error for %s", tc.versionStr)
			} else {
				require.NoError(t, err, "ParseVersion should not error for %s", tc.versionStr)
				assert.GreaterOrEqual(t, major, 0, "Major should be non-negative")
				assert.GreaterOrEqual(t, minor, 0, "Minor should be non-negative")
			}
		})
	}
}

// ============================================================================
// VERSION CAPABILITIES TESTS
// ============================================================================

func TestVersionCapabilities(t *testing.T) {
	testCases := []struct {
		name                  string
		version               int
		hasWriteLagColumns    bool
		hasLogicalReplication bool
		hasWalReceiver        bool
		hasPublication        bool
		hasStandbySignal      bool
		hasPgStatWal          bool
		hasPgStatSubscription bool
	}{
		{"PostgreSQL 11", 11, false, false, true, false, false, false, false},
		{"PostgreSQL 12", 12, false, true, true, true, true, false, true},
		{"PostgreSQL 13", 13, true, true, true, true, true, false, true},
		{"PostgreSQL 14", 14, true, true, true, true, true, true, true},
		{"PostgreSQL 15", 15, true, true, true, true, true, true, true},
		{"PostgreSQL 16", 16, true, true, true, true, true, true, true},
		{"PostgreSQL 17", 17, true, true, true, true, true, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock storage to test GetVersionCapabilities
			// Since it doesn't require DB access, we can pass nil context
			mockStore := &storage.PostgresDB{}

			caps, err := mockStore.GetVersionCapabilities(nil, tc.version)
			require.NoError(t, err, "GetVersionCapabilities should not error")

			assert.Equal(t, tc.hasWriteLagColumns, caps.HasWriteLagColumns, "HasWriteLagColumns mismatch")
			assert.Equal(t, tc.hasLogicalReplication, caps.HasLogicalReplication, "HasLogicalReplication mismatch")
			assert.Equal(t, tc.hasWalReceiver, caps.HasWalReceiver, "HasWalReceiver mismatch")
			assert.Equal(t, tc.hasPublication, caps.HasPublication, "HasPublication mismatch")
			assert.Equal(t, tc.hasStandbySignal, caps.HasStandbySignal, "HasStandbySignal mismatch")
			assert.Equal(t, tc.hasPgStatWal, caps.HasPgStatWal, "HasPgStatWal mismatch")
			assert.Equal(t, tc.hasPgStatSubscription, caps.HasPgStatSubscription, "HasPgStatSubscription mismatch")
		})
	}
}

// ============================================================================
// VERSION SUPPORT STATUS TESTS
// ============================================================================

func TestIsVersionSupported(t *testing.T) {
	testCases := []struct {
		version   int
		supported bool
	}{
		{17, true},
		{16, true},
		{15, true},
		{14, true},
		{13, true},
		{12, false}, // EOL
		{11, false}, // EOL
		{10, false}, // EOL
	}

	for _, tc := range testCases {
		t.Run("PostgreSQL version", func(t *testing.T) {
			result := storage.IsVersionSupported(tc.version)
			assert.Equal(t, tc.supported, result, "IsVersionSupported mismatch for version %d", tc.version)
		})
	}
}

// ============================================================================
// VERSION COMPARISON TESTS
// ============================================================================

func TestVersionComparison(t *testing.T) {
	testCases := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"16.0", "16.0", 0},
		{"16.0", "15.0", 1},
		{"15.0", "16.0", -1},
		{"16.2", "16.1", 1},
		{"16.1", "16.2", -1},
		{"17.0", "11.0", 1},
		{"11.0", "17.0", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.v1+" vs "+tc.v2, func(t *testing.T) {
			result := storage.CompareVersions(tc.v1, tc.v2)
			assert.Equal(t, tc.expected, result, "CompareVersions(%s, %s) mismatch", tc.v1, tc.v2)
		})
	}
}

// ============================================================================
// SUPPORTED VERSIONS LIST TEST
// ============================================================================

func TestGetAllSupportedVersions(t *testing.T) {
	mockStore := &storage.PostgresDB{}
	versions := mockStore.GetAllSupportedVersions()

	assert.GreaterOrEqual(t, len(versions), 7, "Should have at least 7 supported versions")

	// Check that we have expected versions
	versionMap := make(map[int]bool)
	for _, v := range versions {
		versionMap[v.Major] = true
	}

	assert.True(t, versionMap[17], "Should include PostgreSQL 17")
	assert.True(t, versionMap[16], "Should include PostgreSQL 16")
	assert.True(t, versionMap[15], "Should include PostgreSQL 15")
	assert.True(t, versionMap[14], "Should include PostgreSQL 14")
	assert.True(t, versionMap[13], "Should include PostgreSQL 13")
	assert.True(t, versionMap[12], "Should include PostgreSQL 12 (EOL but supported for detection)")
	assert.True(t, versionMap[11], "Should include PostgreSQL 11 (EOL but supported for detection)")

	// Check that support status is correct
	for _, v := range versions {
		if v.Major >= 13 && v.Major <= 17 {
			assert.True(t, v.IsSupported, "PostgreSQL %d should be marked as supported", v.Major)
		} else {
			assert.False(t, v.IsSupported, "PostgreSQL %d should be marked as not supported", v.Major)
		}
	}
}
