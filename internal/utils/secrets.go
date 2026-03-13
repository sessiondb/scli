// Package utils provides secret generation for scli.
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

// GenerateEncryptionKey returns a 32-byte random value as 32 hex characters (16 bytes hex-encoded, truncated to 32 chars per spec).
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key)[:32], nil
}

// GenerateToken returns a URL-safe base64-encoded random token (24 bytes) for MIGRATE_TOKEN.
func GenerateToken() (string, error) {
	token := make([]byte, 24)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}
