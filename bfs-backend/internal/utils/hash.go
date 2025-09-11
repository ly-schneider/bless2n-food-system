package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	ArgonTime       = uint32(3)         // iterations
	ArgonMemory     = uint32(64 * 1024) // 64 MiB
	ArgonThreads    = uint8(1)          // parallelism
	ArgonKeyLen     = uint32(32)        // 32 bytes
	ArgonSaltLength = 16                // 16 bytes
)

type argon2Params struct {
	memory  uint32
	time    uint32
	threads uint8
}

// HashOTPArgon2 hashes an OTP code using Argon2id and returns PHC format string
func HashOTPArgon2(code string) (string, error) {
	salt := make([]byte, ArgonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("salt: %w", err)
	}
	hash := argon2.IDKey([]byte(code), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)

	encSalt := base64.RawStdEncoding.EncodeToString(salt)
	encHash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, ArgonMemory, ArgonTime, ArgonThreads, encSalt, encHash), nil
}

// VerifyOTPArgon2 verifies an OTP code against its Argon2id hash
func VerifyOTPArgon2(code, encoded string) (bool, error) {
	p, salt, expected, err := parsePHC(encoded)
	if err != nil {
		return false, err
	}
	computed := argon2.IDKey([]byte(code), salt, p.time, p.memory, p.threads, uint32(len(expected)))
	return subtle.ConstantTimeCompare(computed, expected) == 1, nil
}

// GenerateOTP generates a 6-digit OTP code
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// GenerateFamilyID generates a random family ID for refresh token rotation
func GenerateFamilyID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate family ID: %w", err)
	}
	return fmt.Sprintf("%x", bytes), nil
}

// parsePHC parses a PHC formatted Argon2 hash string
func parsePHC(encoded string) (p argon2Params, salt, hash []byte, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return p, nil, nil, errors.New("invalid argon2 encoded hash format")
	}
	// version
	if !strings.HasPrefix(parts[2], "v=") {
		return p, nil, nil, errors.New("missing version")
	}
	v, err := strconv.Atoi(strings.TrimPrefix(parts[2], "v="))
	if err != nil || v != argon2.Version {
		return p, nil, nil, fmt.Errorf("unsupported argon2 version: %v", parts[2])
	}
	// params m,t,p
	paramKV := strings.Split(parts[3], ",")
	if len(paramKV) != 3 {
		return p, nil, nil, errors.New("invalid param segment")
	}
	get := func(prefix string) (string, error) {
		for _, kv := range paramKV {
			if value, found := strings.CutPrefix(kv, prefix); found {
				return value, nil
			}
		}
		return "", fmt.Errorf("missing %s", prefix)
	}
	memStr, err := get("m=")
	if err != nil {
		return p, nil, nil, err
	}
	timeStr, err := get("t=")
	if err != nil {
		return p, nil, nil, err
	}
	thStr, err := get("p=")
	if err != nil {
		return p, nil, nil, err
	}
	mem, err := strconv.ParseUint(memStr, 10, 32)
	if err != nil {
		return p, nil, nil, fmt.Errorf("memory: %w", err)
	}
	t, err := strconv.ParseUint(timeStr, 10, 32)
	if err != nil {
		return p, nil, nil, fmt.Errorf("time: %w", err)
	}
	th, err := strconv.ParseUint(thStr, 10, 8)
	if err != nil {
		return p, nil, nil, fmt.Errorf("threads: %w", err)
	}
	p = argon2Params{memory: uint32(mem), time: uint32(t), threads: uint8(th)}

	// salt/hash
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return p, nil, nil, fmt.Errorf("salt b64: %w", err)
	}
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return p, nil, nil, fmt.Errorf("hash b64: %w", err)
	}
	return p, salt, hash, nil
}
