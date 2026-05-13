package query_performance

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// Fingerprinter generates fingerprints for SQL queries
// to group queries with the same structure but different parameters
type Fingerprinter struct {
	logger *zap.Logger
}

// NewFingerprinter creates a new Fingerprinter instance
func NewFingerprinter() *Fingerprinter {
	return &Fingerprinter{
		logger: zap.NewNop(), // Default to no-op logger
	}
}

// NewFingerprinterWithLogger creates a new Fingerprinter with a custom logger
func NewFingerprinterWithLogger(logger *zap.Logger) *Fingerprinter {
	return &Fingerprinter{
		logger: logger,
	}
}

// Patterns for normalizing SQL queries
var (
	// Match quoted strings (single and double quotes)
	stringPattern = regexp.MustCompile(`'(?:[^']|'')*'|"(?:[^"]|"")*"`)

	// Match numeric literals (integers and decimals)
	numberPattern = regexp.MustCompile(`\b\d+(?:\.\d+)?\b`)

	// Match IN lists: IN (1, 2, 3) -> IN (?)
	inListPattern = regexp.MustCompile(`\bIN\s*\(\s*[^)]+\)`)

	// Match VALUES lists for INSERT statements
	valuesPattern = regexp.MustCompile(`VALUES\s*\([^)]+\)`)

	// Match multiple spaces
	multiSpacePattern = regexp.MustCompile(`\s+`)
)

// Fingerprint generates a 32-character hex fingerprint hash for a query
// Queries with the same structure but different parameter values will have the same fingerprint
func (f *Fingerprinter) Fingerprint(queryText string) string {
	// Normalize the query to remove specific parameter values
	normalized := f.normalizeForFingerprint(queryText)

	// Hash the normalized query
	return hashString(normalized)
}

// Normalize returns a parameterized version of the query
// e.g., "SELECT * FROM users WHERE id = 1" -> "SELECT * FROM users WHERE id = $1"
func (f *Fingerprinter) Normalize(queryText string) (string, error) {
	// Normalize the query by replacing literals with placeholders
	normalized := f.normalizeWithPlaceholders(queryText)
	return normalized, nil
}

// normalizeForFingerprint removes literal values from a query for fingerprinting
func (f *Fingerprinter) normalizeForFingerprint(queryText string) string {
	// Convert to uppercase for consistent comparison
	normalized := strings.ToUpper(queryText)

	// Remove string literals
	normalized = stringPattern.ReplaceAllString(normalized, "?")

	// Remove numeric literals
	normalized = numberPattern.ReplaceAllString(normalized, "?")

	// Normalize IN lists
	normalized = inListPattern.ReplaceAllStringFunc(normalized, func(match string) string {
		// Keep the IN keyword but replace the list with (?)
		return "IN (?)"
	})

	// Normalize VALUES lists
	normalized = valuesPattern.ReplaceAllString(normalized, "VALUES (?)")

	// Normalize whitespace
	normalized = multiSpacePattern.ReplaceAllString(normalized, " ")

	// Trim
	normalized = strings.TrimSpace(normalized)

	return normalized
}

// normalizeWithPlaceholders replaces literal values with $1, $2, etc.
func (f *Fingerprinter) normalizeWithPlaceholders(queryText string) string {
	placeholderNum := 0
	normalized := queryText

	// Replace string literals with $N
	normalized = stringPattern.ReplaceAllStringFunc(normalized, func(match string) string {
		placeholderNum++
		return "$" + strconv.Itoa(placeholderNum)
	})

	// Replace numeric literals with $N
	normalized = numberPattern.ReplaceAllStringFunc(normalized, func(match string) string {
		placeholderNum++
		return "$" + strconv.Itoa(placeholderNum)
	})

	// Normalize whitespace
	normalized = multiSpacePattern.ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

// hashString creates a 32-character hex hash from a string
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	// Return first 16 bytes as hex (32 characters)
	return hex.EncodeToString(hash[:16])
}
