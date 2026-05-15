#include "gtest/gtest.h"
#include <string>
#include <vector>
#include <algorithm>

/**
 * Data Classification Pattern Tests
 *
 * Tests for pattern validation functions used in data classification:
 * - CPF (Brazilian Individual Taxpayer Registry) validation
 * - CNPJ (Brazilian National Registry of Legal Entities) validation
 * - Email pattern detection
 * - Phone pattern detection for Brazilian format
 * - Credit card Luhn algorithm validation
 * - Credit card type detection
 *
 * These tests validate the pattern detection algorithms that would be used
 * by a DataClassificationCollector plugin.
 */

// ============================================================================
// CPF Validation Functions
// ============================================================================

/**
 * Validates a CPF number using check digit algorithm
 * CPF format: XXX.XXX.XXX-XX (11 digits)
 *
 * Algorithm:
 * 1. Extract the 9 base digits
 * 2. Calculate first check digit: sum(digit[i] * (10-i)) % 11, if < 2 -> 0, else 11 - result
 * 3. Calculate second check digit: sum(digit[i] * (11-i)) % 11, if < 2 -> 0, else 11 - result
 * 4. Compare calculated digits with provided check digits
 */
bool validateCPF(const std::string& cpf) {
    // Remove non-digit characters
    std::string digits;
    for (char c : cpf) {
        if (std::isdigit(c)) {
            digits += c;
        }
    }

    // CPF must have exactly 11 digits
    if (digits.length() != 11) {
        return false;
    }

    // Check for all same digits (invalid but passes check digit algorithm)
    if (digits.find_first_not_of(digits[0]) == std::string::npos) {
        return false;
    }

    // Calculate first check digit
    int sum = 0;
    for (int i = 0; i < 9; i++) {
        sum += (digits[i] - '0') * (10 - i);
    }
    int remainder = sum % 11;
    int firstCheck = (remainder < 2) ? 0 : 11 - remainder;

    if (firstCheck != (digits[9] - '0')) {
        return false;
    }

    // Calculate second check digit
    sum = 0;
    for (int i = 0; i < 10; i++) {
        sum += (digits[i] - '0') * (11 - i);
    }
    remainder = sum % 11;
    int secondCheck = (remainder < 2) ? 0 : 11 - remainder;

    return secondCheck == (digits[10] - '0');
}

// ============================================================================
// CNPJ Validation Functions
// ============================================================================

/**
 * Validates a CNPJ number using check digit algorithm
 * CNPJ format: XX.XXX.XXX/XXXX-XX (14 digits)
 *
 * Algorithm uses weighted multiplication with specific weights
 */
bool validateCNPJ(const std::string& cnpj) {
    // Remove non-digit characters
    std::string digits;
    for (char c : cnpj) {
        if (std::isdigit(c)) {
            digits += c;
        }
    }

    // CNPJ must have exactly 14 digits
    if (digits.length() != 14) {
        return false;
    }

    // Check for all same digits
    if (digits.find_first_not_of(digits[0]) == std::string::npos) {
        return false;
    }

    // Weights for first check digit
    const std::vector<int> weights1 = {5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2};

    // Calculate first check digit
    int sum = 0;
    for (int i = 0; i < 12; i++) {
        sum += (digits[i] - '0') * weights1[i];
    }
    int remainder = sum % 11;
    int firstCheck = (remainder < 2) ? 0 : 11 - remainder;

    if (firstCheck != (digits[12] - '0')) {
        return false;
    }

    // Weights for second check digit
    const std::vector<int> weights2 = {6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2};

    // Calculate second check digit
    sum = 0;
    for (int i = 0; i < 13; i++) {
        sum += (digits[i] - '0') * weights2[i];
    }
    remainder = sum % 11;
    int secondCheck = (remainder < 2) ? 0 : 11 - remainder;

    return secondCheck == (digits[13] - '0');
}

// ============================================================================
// Luhn Algorithm for Credit Cards
// ============================================================================

/**
 * Validates a credit card number using the Luhn algorithm
 *
 * Algorithm:
 * 1. From rightmost digit, double every second digit
 * 2. If doubling results in > 9, subtract 9
 * 3. Sum all digits
 * 4. Valid if sum % 10 == 0
 */
bool validateLuhn(const std::string& number) {
    // Remove non-digit characters
    std::string digits;
    for (char c : number) {
        if (std::isdigit(c)) {
            digits += c;
        }
    }

    if (digits.empty()) {
        return false;
    }

    int sum = 0;
    bool doubleDigit = false;

    // Process from right to left
    for (int i = digits.length() - 1; i >= 0; i--) {
        int digit = digits[i] - '0';

        if (doubleDigit) {
            digit *= 2;
            if (digit > 9) {
                digit -= 9;
            }
        }

        sum += digit;
        doubleDigit = !doubleDigit;
    }

    return sum % 10 == 0;
}

/**
 * Detects credit card type based on number prefix
 */
std::string detectCardType(const std::string& number) {
    // Remove non-digit characters
    std::string digits;
    for (char c : number) {
        if (std::isdigit(c)) {
            digits += c;
        }
    }

    if (digits.length() < 2) {
        return "UNKNOWN";
    }

    // Visa: starts with 4
    if (digits[0] == '4') {
        return "VISA";
    }

    // Mastercard: starts with 51-55 or 2221-2720
    if (digits.length() >= 2) {
        int prefix = std::stoi(digits.substr(0, 2));
        if (prefix >= 51 && prefix <= 55) {
            return "MASTERCARD";
        }
    }
    if (digits.length() >= 4) {
        int prefix = std::stoi(digits.substr(0, 4));
        if (prefix >= 2221 && prefix <= 2720) {
            return "MASTERCARD";
        }
    }

    // American Express: starts with 34 or 37
    if (digits.substr(0, 2) == "34" || digits.substr(0, 2) == "37") {
        return "AMEX";
    }

    // Discover: starts with 6011, 644-649, or 65
    if (digits.substr(0, 4) == "6011") {
        return "DISCOVER";
    }
    if (digits.length() >= 3) {
        int prefix3 = std::stoi(digits.substr(0, 3));
        if (prefix3 >= 644 && prefix3 <= 649) {
            return "DISCOVER";
        }
    }
    if (digits.substr(0, 2) == "65") {
        return "DISCOVER";
    }

    // Diners Club: starts with 300-305, 36, or 38
    if (digits.length() >= 3) {
        int prefix3 = std::stoi(digits.substr(0, 3));
        if (prefix3 >= 300 && prefix3 <= 305) {
            return "DINERS_CLUB";
        }
    }
    if (digits.substr(0, 2) == "36" || digits.substr(0, 2) == "38") {
        return "DINERS_CLUB";
    }

    return "UNKNOWN";
}

// ============================================================================
// Email Validation
// ============================================================================

/**
 * Simple email format validation
 * Checks for basic structure: local@domain
 */
bool validateEmail(const std::string& email) {
    // Must contain @
    size_t atPos = email.find('@');
    if (atPos == std::string::npos || atPos == 0 || atPos == email.length() - 1) {
        return false;
    }

    // Must have a domain with at least one dot
    std::string domain = email.substr(atPos + 1);
    if (domain.find('.') == std::string::npos) {
        return false;
    }

    // Basic character validation
    for (char c : email) {
        if (!std::isalnum(c) && c != '.' && c != '@' && c != '_' && c != '-' && c != '+') {
            return false;
        }
    }

    return true;
}

// ============================================================================
// Brazilian Phone Validation
// ============================================================================

/**
 * Validates Brazilian phone number format
 * Formats: (XX) XXXXX-XXXX or (XX) XXXX-XXXX
 */
bool validateBrazilianPhone(const std::string& phone) {
    // Remove non-digit characters
    std::string digits;
    for (char c : phone) {
        if (std::isdigit(c)) {
            digits += c;
        }
    }

    // Brazilian phone numbers have 10 or 11 digits (including area code)
    if (digits.length() != 10 && digits.length() != 11) {
        return false;
    }

    // Area code must be 11-99 (valid Brazilian area codes)
    int areaCode = std::stoi(digits.substr(0, 2));
    if (areaCode < 11 || areaCode > 99) {
        return false;
    }

    return true;
}

// ============================================================================
// Tests
// ============================================================================

class DataClassificationTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Setup test fixtures if needed
    }
};

// ============================================================================
// CPF Tests
// ============================================================================

TEST_F(DataClassificationTest, ValidateCPFWithValidNumber) {
    // Valid CPF with correct check digits: 123.456.789-09
    EXPECT_TRUE(validateCPF("123.456.789-09"));
    EXPECT_TRUE(validateCPF("12345678909"));  // Without formatting
}

TEST_F(DataClassificationTest, ValidateCPFWithInvalidNumber) {
    // Invalid CPF with wrong check digits
    EXPECT_FALSE(validateCPF("123.456.789-00"));
    EXPECT_FALSE(validateCPF("123.456.789-99"));
}

TEST_F(DataClassificationTest, ValidateCPFRejectsAllSameDigits) {
    // All same digits should be rejected (passes algorithm but is invalid)
    EXPECT_FALSE(validateCPF("111.111.111-11"));
    EXPECT_FALSE(validateCPF("222.222.222-22"));
    EXPECT_FALSE(validateCPF("000.000.000-00"));
}

TEST_F(DataClassificationTest, ValidateCPFRejectsInvalidLength) {
    EXPECT_FALSE(validateCPF("123.456.789"));     // Too few digits
    EXPECT_FALSE(validateCPF("123.456.789-0001")); // Too many digits
    EXPECT_FALSE(validateCPF(""));                  // Empty
}

TEST_F(DataClassificationTest, ValidateCPFHandlesFormatting) {
    // Valid CPF in different formats
    EXPECT_TRUE(validateCPF("123.456.789-09"));
    EXPECT_TRUE(validateCPF("12345678909"));
    EXPECT_TRUE(validateCPF("123 456 789 09"));
}

// ============================================================================
// CNPJ Tests
// ============================================================================

TEST_F(DataClassificationTest, ValidateCNPJWithValidNumber) {
    // Valid CNPJ: 11.222.333/0001-81
    EXPECT_TRUE(validateCNPJ("11.222.333/0001-81"));
    EXPECT_TRUE(validateCNPJ("11222333000181"));  // Without formatting
}

TEST_F(DataClassificationTest, ValidateCNPJWithInvalidNumber) {
    // Invalid CNPJ with wrong check digits
    EXPECT_FALSE(validateCNPJ("11.222.333/0001-00"));
    EXPECT_FALSE(validateCNPJ("11.222.333/0001-99"));
}

TEST_F(DataClassificationTest, ValidateCNPJRejectsAllSameDigits) {
    // All same digits should be rejected
    EXPECT_FALSE(validateCNPJ("11.111.111/1111-11"));
    EXPECT_FALSE(validateCNPJ("00.000.000/0000-00"));
}

TEST_F(DataClassificationTest, ValidateCNPJRejectsInvalidLength) {
    EXPECT_FALSE(validateCNPJ("11.222.333/0001"));     // Too few digits
    EXPECT_FALSE(validateCNPJ("11.222.333/0001-8100")); // Too many digits
}

// ============================================================================
// Luhn Algorithm Tests
// ============================================================================

TEST_F(DataClassificationTest, ValidateCreditCardWithLuhn) {
    // Valid test card numbers (pass Luhn check)
    EXPECT_TRUE(validateLuhn("4111111111111111"));  // Visa test number
    EXPECT_TRUE(validateLuhn("5500000000000004"));  // Mastercard test number
    EXPECT_TRUE(validateLuhn("340000000000009"));   // Amex test number
}

TEST_F(DataClassificationTest, RejectInvalidCreditCardWithLuhn) {
    // Invalid card numbers (fail Luhn check)
    EXPECT_FALSE(validateLuhn("4111111111111112"));  // Last digit changed
    EXPECT_FALSE(validateLuhn("5500000000000005"));  // Last digit changed
    EXPECT_FALSE(validateLuhn("1234567890123456"));  // Random invalid
}

TEST_F(DataClassificationTest, ValidateLuhnHandlesFormatting) {
    // Should work with spaces and dashes
    EXPECT_TRUE(validateLuhn("4111-1111-1111-1111"));
    EXPECT_TRUE(validateLuhn("4111 1111 1111 1111"));
}

TEST_F(DataClassificationTest, ValidateLuhnRejectsEmpty) {
    EXPECT_FALSE(validateLuhn(""));
    EXPECT_FALSE(validateLuhn("   "));
}

// ============================================================================
// Card Type Detection Tests
// ============================================================================

TEST_F(DataClassificationTest, DetectCardTypeVisa) {
    EXPECT_EQ(detectCardType("4111111111111111"), "VISA");
    EXPECT_EQ(detectCardType("4000000000000002"), "VISA");
}

TEST_F(DataClassificationTest, DetectCardTypeMastercard) {
    EXPECT_EQ(detectCardType("5500000000000004"), "MASTERCARD");  // 51-55 range
    EXPECT_EQ(detectCardType("5200000000000007"), "MASTERCARD");  // 51-55 range
    EXPECT_EQ(detectCardType("2221000000000009"), "MASTERCARD");  // 2221-2720 range
}

TEST_F(DataClassificationTest, DetectCardTypeAmex) {
    EXPECT_EQ(detectCardType("340000000000009"), "AMEX");  // 34 prefix
    EXPECT_EQ(detectCardType("370000000000002"), "AMEX");  // 37 prefix
}

TEST_F(DataClassificationTest, DetectCardTypeDiscover) {
    EXPECT_EQ(detectCardType("6011000000000004"), "DISCOVER");  // 6011 prefix
    EXPECT_EQ(detectCardType("6500000000000002"), "DISCOVER");  // 65 prefix
}

TEST_F(DataClassificationTest, DetectCardTypeUnknown) {
    EXPECT_EQ(detectCardType("0000000000000000"), "UNKNOWN");
    EXPECT_EQ(detectCardType("1234567890123456"), "UNKNOWN");
}

// ============================================================================
// Email Validation Tests
// ============================================================================

TEST_F(DataClassificationTest, ValidateEmailWithValidAddress) {
    EXPECT_TRUE(validateEmail("test@example.com"));
    EXPECT_TRUE(validateEmail("user.name@domain.org"));
    EXPECT_TRUE(validateEmail("user+tag@company.co.uk"));
    EXPECT_TRUE(validateEmail("user_name@domain.com"));
    EXPECT_TRUE(validateEmail("user-name@domain.com"));
}

TEST_F(DataClassificationTest, ValidateEmailRejectsInvalidAddress) {
    EXPECT_FALSE(validateEmail("invalid"));              // No @
    EXPECT_FALSE(validateEmail("@domain.com"));           // No local part
    EXPECT_FALSE(validateEmail("user@"));                 // No domain
    EXPECT_FALSE(validateEmail("user@domain"));           // No TLD
    EXPECT_FALSE(validateEmail("user name@domain.com"));  // Space in local part
}

// ============================================================================
// Phone Validation Tests
// ============================================================================

TEST_F(DataClassificationTest, ValidateBrazilianPhoneWithValidNumber) {
    // Mobile format: (XX) XXXXX-XXXX (11 digits)
    EXPECT_TRUE(validateBrazilianPhone("(11) 99999-9999"));
    EXPECT_TRUE(validateBrazilianPhone("11999999999"));

    // Landline format: (XX) XXXX-XXXX (10 digits)
    EXPECT_TRUE(validateBrazilianPhone("(11) 3333-3333"));
    EXPECT_TRUE(validateBrazilianPhone("1133333333"));
}

TEST_F(DataClassificationTest, ValidateBrazilianPhoneRejectsInvalidNumber) {
    // Invalid area code
    EXPECT_FALSE(validateBrazilianPhone("(01) 9999-9999"));  // Area code 01 invalid
    EXPECT_FALSE(validateBrazilianPhone("(10) 9999-9999"));  // Area code 10 invalid

    // Wrong length
    EXPECT_FALSE(validateBrazilianPhone("119999999"));   // Too few digits
    EXPECT_FALSE(validateBrazilianPhone("119999999999")); // Too many digits
}

TEST_F(DataClassificationTest, ValidateBrazilianPhoneHandlesFormatting) {
    // Should work with different formatting
    EXPECT_TRUE(validateBrazilianPhone("11999999999"));
    EXPECT_TRUE(validateBrazilianPhone("(11) 99999-9999"));
    EXPECT_TRUE(validateBrazilianPhone("11 99999 9999"));
}