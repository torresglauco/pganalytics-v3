package unit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
)

// TestComputeCertificateThumbprint tests the SHA256 thumbprint computation
func TestComputeCertificateThumbprint(t *testing.T) {
	// Generate a test certificate
	certDER := generateTestCertificateDER(t)

	// Compute thumbprint
	thumbprint := auth.ComputeCertificateThumbprint(certDER)

	// Verify it's a valid hex string
	if len(thumbprint) != 64 {
		t.Errorf("Expected thumbprint length 64, got %d", len(thumbprint))
	}

	// Verify it's valid hex
	if _, err := hex.DecodeString(thumbprint); err != nil {
		t.Errorf("Thumbprint is not valid hex: %v", err)
	}

	// Verify it's consistent
	thumbprint2 := auth.ComputeCertificateThumbprint(certDER)
	if thumbprint != thumbprint2 {
		t.Errorf("Thumbprints don't match: %s != %s", thumbprint, thumbprint2)
	}
}

// TestComputeCertificateThumbprintIsUnique tests that different certs have different thumbprints
func TestComputeCertificateThumbprintIsUnique(t *testing.T) {
	cert1 := generateTestCertificateDER(t)
	cert2 := generateTestCertificateDER(t)

	thumbprint1 := auth.ComputeCertificateThumbprint(cert1)
	thumbprint2 := auth.ComputeCertificateThumbprint(cert2)

	if thumbprint1 == thumbprint2 {
		t.Errorf("Different certificates should have different thumbprints")
	}
}

// TestComputeCertificateThumbprintNotEmpty tests that thumbprint is not empty
func TestComputeCertificateThumbprintNotEmpty(t *testing.T) {
	certDER := generateTestCertificateDER(t)
	thumbprint := auth.ComputeCertificateThumbprint(certDER)

	if thumbprint == "" {
		t.Errorf("Thumbprint should not be empty")
	}
}

// TestComputeCertificateThumbprintUsesActualSHA256 tests that we compute actual SHA256, not a placeholder
func TestComputeCertificateThumbprintUsesActualSHA256(t *testing.T) {
	testData := []byte("test certificate data")

	// Compute SHA256 manually
	expectedHash := sha256.Sum256(testData)
	expectedThumbprint := fmt.Sprintf("%x", expectedHash)

	// Compute using the function
	actualThumbprint := auth.ComputeCertificateThumbprint(testData)

	if actualThumbprint != expectedThumbprint {
		t.Errorf("Expected %s, got %s", expectedThumbprint, actualThumbprint)
	}

	// Verify it's not a hardcoded placeholder like "demo-thumbprint-hash"
	if actualThumbprint == "demo-thumbprint-hash-32chars123" {
		t.Errorf("Thumbprint appears to be a hardcoded placeholder")
	}
}

// TestGenerateCollectorCertificateThumbprint tests that generated certificates have valid thumbprints
func TestGenerateCollectorCertificateThumbprint(t *testing.T) {
	cm, err := auth.NewCertificateManager("", "")
	if err != nil {
		t.Fatalf("Failed to create certificate manager: %v", err)
	}

	collectorID := uuid.New()
	hostname := "test-collector.example.com"
	validityDays := 365

	certPair, err := cm.GenerateCollectorCertificate(collectorID, hostname, validityDays)
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// Verify thumbprint is not empty
	if certPair.Thumbprint == "" {
		t.Errorf("Generated certificate should have a thumbprint")
	}

	// Verify thumbprint is valid hex with correct length
	if len(certPair.Thumbprint) != 64 {
		t.Errorf("Expected thumbprint length 64, got %d", len(certPair.Thumbprint))
	}

	if _, err := hex.DecodeString(certPair.Thumbprint); err != nil {
		t.Errorf("Thumbprint is not valid hex: %v", err)
	}

	// Verify it's not a placeholder
	if certPair.Thumbprint == "demo-thumbprint-hash-32chars123" {
		t.Errorf("Certificate thumbprint is a placeholder, not actual SHA256 hash")
	}
}

// TestGenerateCollectorCertificateConsistentThumbprint tests that regenerating from same cert gives same thumbprint
func TestGenerateCollectorCertificateConsistentThumbprint(t *testing.T) {
	cm, err := auth.NewCertificateManager("", "")
	if err != nil {
		t.Fatalf("Failed to create certificate manager: %v", err)
	}

	collectorID := uuid.New()
	hostname := "test-collector.example.com"

	certPair1, err := cm.GenerateCollectorCertificate(collectorID, hostname, 365)
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// Extract the certificate from PEM
	block, _ := pem.Decode([]byte(certPair1.Certificate))
	if block == nil {
		t.Fatalf("Failed to parse certificate PEM")
	}

	// Recompute thumbprint from DER bytes
	recomputedThumbprint := auth.ComputeCertificateThumbprint(block.Bytes)

	if recomputedThumbprint != certPair1.Thumbprint {
		t.Errorf("Recomputed thumbprint doesn't match: %s != %s", recomputedThumbprint, certPair1.Thumbprint)
	}
}

// TestThumbprintFormatValid tests that thumbprints are in correct format
func TestThumbprintFormatValid(t *testing.T) {
	cm, err := auth.NewCertificateManager("", "")
	if err != nil {
		t.Fatalf("Failed to create certificate manager: %v", err)
	}

	for i := 0; i < 5; i++ {
		collectorID := uuid.New()
		hostname := fmt.Sprintf("test-collector-%d.example.com", i)

		certPair, err := cm.GenerateCollectorCertificate(collectorID, hostname, 365)
		if err != nil {
			t.Fatalf("Failed to generate certificate: %v", err)
		}

		// Verify format: should be 64 hex characters
		if len(certPair.Thumbprint) != 64 {
			t.Errorf("Certificate %d: Expected thumbprint length 64, got %d", i, len(certPair.Thumbprint))
		}

		// Verify all characters are valid hex
		for j, c := range certPair.Thumbprint {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("Certificate %d: Invalid hex character at position %d: %c", i, j, c)
			}
		}
	}
}

// Helper function to generate test certificate DER
func generateTestCertificateDER(t *testing.T) []byte {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "test-collector.example.com",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	return certDER
}
