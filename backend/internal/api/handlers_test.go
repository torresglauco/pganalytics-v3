package api

import (
	"testing"
)

// TestGenerateTemporaryPassword tests secure password generation
func TestGenerateTemporaryPassword(t *testing.T) {
	// Generate multiple passwords to ensure diversity
	passwords := make(map[string]bool)

	for i := 0; i < 100; i++ {
		pwd := generateTemporaryPassword(12)

		// Check password length
		if len(pwd) != 12 {
			t.Errorf("Expected password length 12, got %d", len(pwd))
		}

		// Check for character variety
		hasUpper := false
		hasLower := false
		hasDigit := false

		for _, c := range pwd {
			if c >= 'A' && c <= 'Z' {
				hasUpper = true
			}
			if c >= 'a' && c <= 'z' {
				hasLower = true
			}
			if c >= '0' && c <= '9' {
				hasDigit = true
			}
		}

		if !hasUpper {
			t.Error("Password missing uppercase character")
			break
		}
		if !hasLower {
			t.Error("Password missing lowercase character")
			break
		}
		if !hasDigit {
			t.Error("Password missing digit")
			break
		}

		passwords[pwd] = true
	}

	// Check for diversity (most passwords should be different)
	if len(passwords) < 80 { // Allow some slight variance in randomness
		t.Errorf("Generated passwords lack sufficient diversity: %d unique out of 100", len(passwords))
	}
}

// TestSecureRandInt tests the secure random integer generation
func TestSecureRandInt(t *testing.T) {
	// Test that secureRandInt returns values in the expected range
	for n := 1; n <= 100; n++ {
		for i := 0; i < 100; i++ {
			val, err := secureRandInt(n)
			if err != nil {
				t.Fatalf("secureRandInt(%d) returned error: %v", n, err)
			}
			if val < 0 || val >= n {
				t.Errorf("secureRandInt(%d) returned %d, expected value in [0, %d)", n, val, n)
			}
		}
	}
}

// TestSecureRandIntEdgeCases tests edge cases for secureRandInt
func TestSecureRandIntEdgeCases(t *testing.T) {
	// Test with n=1 (should always return 0)
	for i := 0; i < 10; i++ {
		val, err := secureRandInt(1)
		if err != nil {
			t.Fatalf("secureRandInt(1) returned error: %v", err)
		}
		if val != 0 {
			t.Errorf("secureRandInt(1) returned %d, expected 0", val)
		}
	}

	// Test with invalid input
	_, err := secureRandInt(0)
	if err == nil {
		t.Error("secureRandInt(0) should return error")
	}

	_, err = secureRandInt(-1)
	if err == nil {
		t.Error("secureRandInt(-1) should return error")
	}
}

// TestPasswordCharacterComposition tests password has required character sets
func TestPasswordCharacterComposition(t *testing.T) {
	const (
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		digits    = "0123456789"
	)

	for length := 3; length <= 20; length++ {
		pwd := generateTemporaryPassword(length)

		if len(pwd) != length {
			t.Errorf("Expected password length %d, got %d", length, len(pwd))
		}

		// Verify all characters are in the valid set
		validChars := uppercase + lowercase + digits
		for _, c := range pwd {
			found := false
			for _, valid := range validChars {
				if rune(c) == valid {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Password contains invalid character: %c", c)
			}
		}
	}
}
