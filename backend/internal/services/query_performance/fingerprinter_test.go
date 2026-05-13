package query_performance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFingerprint_SameQueryDifferentParams(t *testing.T) {
	fp := NewFingerprinter()

	query1 := "SELECT * FROM users WHERE id = 1"
	query2 := "SELECT * FROM users WHERE id = 2"

	fingerprint1 := fp.Fingerprint(query1)
	fingerprint2 := fp.Fingerprint(query2)

	assert.Equal(t, fingerprint1, fingerprint2, "Queries with different numeric parameters should have same fingerprint")
	assert.Len(t, fingerprint1, 32, "Fingerprint should be 32 hex characters")
}

func TestFingerprint_DifferentQueries(t *testing.T) {
	fp := NewFingerprinter()

	query1 := "SELECT * FROM users WHERE id = 1"
	query2 := "SELECT * FROM orders WHERE id = 1"

	fingerprint1 := fp.Fingerprint(query1)
	fingerprint2 := fp.Fingerprint(query2)

	assert.NotEqual(t, fingerprint1, fingerprint2, "Different queries should have different fingerprints")
}

func TestFingerprint_Normalize(t *testing.T) {
	fp := NewFingerprinter()

	query := "SELECT * FROM users WHERE id = 1"
	normalized, err := fp.Normalize(query)

	assert.NoError(t, err)
	// Should replace the literal 1 with $1
	assert.True(t, strings.Contains(normalized, "$"), "Normalized query should contain $ placeholder")
	assert.Contains(t, strings.ToUpper(normalized), "SELECT")
	assert.Contains(t, strings.ToUpper(normalized), "USERS")
}

func TestFingerprint_StringLiterals(t *testing.T) {
	fp := NewFingerprinter()

	query1 := "SELECT * FROM users WHERE name = 'Alice'"
	query2 := "SELECT * FROM users WHERE name = 'Bob'"

	fingerprint1 := fp.Fingerprint(query1)
	fingerprint2 := fp.Fingerprint(query2)

	assert.Equal(t, fingerprint1, fingerprint2, "Queries with different string parameters should have same fingerprint")
}

func TestFingerprint_InList(t *testing.T) {
	fp := NewFingerprinter()

	query1 := "SELECT * FROM users WHERE id IN (1, 2, 3)"
	query2 := "SELECT * FROM users WHERE id IN (4, 5, 6)"

	fingerprint1 := fp.Fingerprint(query1)
	fingerprint2 := fp.Fingerprint(query2)

	assert.Equal(t, fingerprint1, fingerprint2, "Queries with different IN list values should have same fingerprint")
}

func TestFingerprint_InsertValues(t *testing.T) {
	fp := NewFingerprinter()

	query1 := "INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com')"
	query2 := "INSERT INTO users (name, email) VALUES ('Bob', 'bob@example.com')"

	fingerprint1 := fp.Fingerprint(query1)
	fingerprint2 := fp.Fingerprint(query2)

	assert.Equal(t, fingerprint1, fingerprint2, "INSERT queries with different values should have same fingerprint")
}

func TestFingerprint_ComplexQuery(t *testing.T) {
	fp := NewFingerprinter()

	query1 := `
		SELECT u.name, COUNT(o.id) as order_count
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.created_at > '2024-01-01'
		AND u.status = 'active'
		GROUP BY u.id
		HAVING COUNT(o.id) > 5
	`

	query2 := `
		SELECT u.name, COUNT(o.id) as order_count
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.created_at > '2024-06-15'
		AND u.status = 'pending'
		GROUP BY u.id
		HAVING COUNT(o.id) > 10
	`

	fingerprint1 := fp.Fingerprint(query1)
	fingerprint2 := fp.Fingerprint(query2)

	assert.Equal(t, fingerprint1, fingerprint2, "Complex queries with different parameters should have same fingerprint")
}

func TestFingerprint_Consistency(t *testing.T) {
	fp := NewFingerprinter()

	query := "SELECT * FROM users WHERE id = 123 AND name = 'Test'"

	// Multiple calls should produce the same fingerprint
	fingerprint1 := fp.Fingerprint(query)
	fingerprint2 := fp.Fingerprint(query)

	assert.Equal(t, fingerprint1, fingerprint2, "Same query should always produce same fingerprint")
}
