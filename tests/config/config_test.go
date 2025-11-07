package config

import (
	"encoding/base64"
	"testing"

	"github.com/go-gost/gost.plus/config"
	"github.com/stretchr/testify/assert"
)

func encodePasswordForTest(password string) string {
	return base64.StdEncoding.EncodeToString([]byte(password))
}

func TestPassword_Operations(t *testing.T) {
	testCases := []struct {
		name              string
		password          string // Input password (already base64 encoded for test cases)
		setPassword       string // Password to set (will be base64 encoded by Set)
		expectEmpty       bool
		expectString      string // Expected result of String()
		expectEncoded     bool   // Whether the password should be base64 encoded after Set
		expectValidBase64 bool   // Whether the input is valid base64
	}{
		// Empty password cases
		{
			name:              "Empty password",
			password:          "",
			setPassword:       "",
			expectEmpty:       true,
			expectString:      "",
			expectEncoded:     false,
			expectValidBase64: true,
		},

		// Simple password cases
		{
			name:              "Simple password",
			password:          encodePasswordForTest("password"),
			setPassword:       "password",
			expectEmpty:       false,
			expectString:      "password",
			expectEncoded:     true,
			expectValidBase64: true,
		},

		// Special characters
		{
			name:              "Special characters",
			password:          encodePasswordForTest("p@ssw0rd!#$%^&*()"),
			setPassword:       "p@ssw0rd!#$%^&*()",
			expectEmpty:       false,
			expectString:      "p@ssw0rd!#$%^&*()",
			expectEncoded:     true,
			expectValidBase64: true,
		},

		// Unicode characters
		{
			name:              "Unicode characters",
			password:          encodePasswordForTest("hello世界"),
			setPassword:       "hello世界",
			expectEmpty:       false,
			expectString:      "hello世界",
			expectEncoded:     true,
			expectValidBase64: true,
		},

		// Password with spaces
		{
			name:              "Password with spaces",
			password:          encodePasswordForTest("my secret password"),
			setPassword:       "my secret password",
			expectEmpty:       false,
			expectString:      "my secret password",
			expectEncoded:     true,
			expectValidBase64: true,
		},

		// Very long password
		{
			name:              "Very long password",
			password:          encodePasswordForTest("this_is_a_very_long_password_that_should_be_encoded_and_decoded_correctly_without_any_issues_whatsoever"),
			setPassword:       "this_is_a_very_long_password_that_should_be_encoded_and_decoded_correctly_without_any_issues_whatsoever",
			expectEmpty:       false,
			expectString:      "this_is_a_very_long_password_that_should_be_encoded_and_decoded_correctly_without_any_issues_whatsoever",
			expectEncoded:     true,
			expectValidBase64: true,
		},

		// Invalid base64 (should be returned as-is by String())
		{
			name:              "Invalid base64",
			password:          "invalid-base64!",
			setPassword:       "invalid-base64!",
			expectEmpty:       false,
			expectString:      "invalid-base64!",
			expectEncoded:     true,
			expectValidBase64: false,
		},

		// Numeric password
		{
			name:              "Numeric password",
			password:          encodePasswordForTest("1234567890"),
			setPassword:       "1234567890",
			expectEmpty:       false,
			expectString:      "1234567890",
			expectEncoded:     true,
			expectValidBase64: true,
		},

		// Pre-encoded base64
		{
			name:              "Pre-encoded base64",
			password:          "bXlTZWNyZXRQYXNzd29yZA==", // "mySecretPassword"
			setPassword:       "mySecretPassword",
			expectEmpty:       false,
			expectString:      "mySecretPassword",
			expectEncoded:     true,
			expectValidBase64: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test direct initialization
			password := config.Password(tc.password)

			// Test String() method
			// For valid base64, it should decode to the original string
			// For invalid base64, it should return the original string
			assert.Equal(t, tc.expectString, password.Reveal(),
				"String() should return the expected value")

			// Test IsEmpty() method
			assert.Equal(t, tc.expectEmpty, password.IsEmpty(),
				"IsEmpty() should return %v for password: %s", tc.expectEmpty, tc.password)

			// Test Set() method
			var setPassword config.Password
			setPassword.Set(tc.setPassword)

			// Verify Set() encodes the password
			if tc.setPassword != "" && tc.expectEncoded {
				assert.NotEqual(t, config.Password(tc.setPassword), setPassword,
					"Set() should encode the password")
			}

			// Verify String() after Set()
			assert.Equal(t, tc.setPassword, setPassword.Reveal(),
				"String() after Set() should return the original password")

			// Additional test for valid base64 handling
			if tc.expectValidBase64 {
				encoded := encodePasswordForTest(tc.expectString)
				assert.Equal(t, tc.expectString, config.Password(encoded).Reveal(),
					"Should correctly decode valid base64 password")
			}
		})
	}
}
