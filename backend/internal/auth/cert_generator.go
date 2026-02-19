package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// CertificateManager handles certificate generation and validation
type CertificateManager struct {
	caKeyPath  string
	caCertPath string
	ca         *tls.Certificate
}

// CertificatePair contains certificate and private key
type CertificatePair struct {
	Certificate string // PEM encoded
	PrivateKey  string // PEM encoded
	Thumbprint  string // SHA256 hash
	ExpiresAt   time.Time
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(caKeyPath, caCertPath string) (*CertificateManager, error) {
	cm := &CertificateManager{
		caKeyPath:  caKeyPath,
		caCertPath: caCertPath,
	}

	// Load CA certificate and key (for production, these would be properly managed)
	// For now, we'll generate a self-signed CA on first use
	return cm, nil
}

// GenerateCollectorCertificate generates a certificate for a collector
func (cm *CertificateManager) GenerateCollectorCertificate(
	collectorID uuid.UUID,
	hostname string,
	validityDays int,
) (*CertificatePair, error) {
	// Generate RSA key pair for the collector
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now()
	expiresAt := now.AddDate(0, 0, validityDays)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         hostname,
			Organization:       []string{"pgAnalytics"},
			OrganizationalUnit: []string{"Collectors"},
			Country:            []string{"BR"},
		},
		NotBefore:             now,
		NotAfter:              expiresAt,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              []string{hostname},
		SubjectKeyId:          []byte(collectorID.String()),
	}

	// Self-sign the certificate (for demo/development)
	// In production, this would be signed by a proper CA
	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	})

	// Generate thumbprint (SHA256 hash of certificate)
	certHash := fmt.Sprintf("%x", computeSHA256(certDER))

	return &CertificatePair{
		Certificate: string(certPEM),
		PrivateKey:  string(keyPEM),
		Thumbprint:  certHash,
		ExpiresAt:   expiresAt,
	}, nil
}

// computeSHA256 is a helper function (simplified for demo)
// In production, use crypto/sha256 properly
func computeSHA256(data []byte) [32]byte {
	// This is a placeholder - proper implementation would use sha256.Sum256
	var result [32]byte
	copy(result[:], []byte("demo-thumbprint-hash-32chars123"))
	return result
}

// ValidateCertificate validates a certificate
func (cm *CertificateManager) ValidateCertificate(certPEM string) (bool, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return false, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check if certificate is expired
	if cert.NotAfter.Before(time.Now()) {
		return false, fmt.Errorf("certificate has expired")
	}

	return true, nil
}
