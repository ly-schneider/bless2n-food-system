package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Config holds the configuration for Argon2 hashing
type Argon2Config struct {
	Time    uint32 // Number of iterations
	Memory  uint32 // Memory usage in KiB
	Threads uint8  // Number of threads
	KeyLen  uint32 // Length of the derived key
	SaltLen uint32 // Length of the salt
}

// DefaultArgon2Config returns OWASP recommended Argon2id parameters
// Time: 1, Memory: 46MB, Threads: 1, KeyLen: 32 bytes, SaltLen: 16 bytes
func DefaultArgon2Config() *Argon2Config {
	return &Argon2Config{
		Time:    1,
		Memory:  47104, // 46MB in KiB
		Threads: 1,
		KeyLen:  32,
		SaltLen: 16,
	}
}

// HashSensitiveData creates an Argon2id hash of sensitive data (passwords, tokens, etc.)
// Returns hash in standard Argon2 encoded format: $argon2id$v=19$m=47104,t=1,p=1$salt$hash
func HashSensitiveData(data string) (string, error) {
	config := DefaultArgon2Config()
	return HashSensitiveDataWithConfig(data, config)
}

// HashSensitiveDataWithConfig creates an Argon2id hash with custom configuration
func HashSensitiveDataWithConfig(data string, config *Argon2Config) (string, error) {
	// Generate random salt
	salt := make([]byte, config.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate hash using Argon2id
	hash := argon2.IDKey([]byte(data), salt, config.Time, config.Memory, config.Threads, config.KeyLen)

	// Encode salt and hash using base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return in standard Argon2 format
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, config.Memory, config.Time, config.Threads, b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifySensitiveData verifies data against an Argon2id hash
// Supports standard Argon2 encoded format: $argon2id$v=19$m=47104,t=1,p=1$salt$hash
func VerifySensitiveData(data, encodedHash string) bool {
	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	// Verify it's Argon2id
	if parts[1] != "argon2id" {
		return false
	}

	// Parse parameters
	var memory, time, threads uint32
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	// Decode stored hash
	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	// Hash the provided data with the same parameters
	dataHash := argon2.IDKey([]byte(data), salt, time, memory, uint8(threads), uint32(len(storedHash)))

	// Compare hashes using constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(storedHash, dataHash) == 1
}

// HashPassword is a convenience function for hashing passwords
func HashPassword(password string) (string, error) {
	return HashSensitiveData(password)
}

// VerifyPassword is a convenience function for verifying passwords
func VerifyPassword(password, hash string) bool {
	return VerifySensitiveData(password, hash)
}

// HashOTP is a convenience function for hashing OTP codes
func HashOTP(otp string) (string, error) {
	return HashSensitiveData(otp)
}

// VerifyOTP is a convenience function for verifying OTP codes
func VerifyOTP(otp, hash string) bool {
	return VerifySensitiveData(otp, hash)
}

// HashToken is a convenience function for hashing tokens (password reset, verification, etc.)
func HashToken(token string) (string, error) {
	return HashSensitiveData(token)
}

// VerifyToken is a convenience function for verifying tokens
func VerifyToken(token, hash string) bool {
	return VerifySensitiveData(token, hash)
}