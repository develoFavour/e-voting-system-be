package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"

	"github.com/develoFavour/e-voting-system-be/internal/models"
)

// EncryptVote encrypts vote data using AES-256
func EncryptVote(voteData *models.VoteData, key string) (string, error) {
	// Convert vote data to JSON
	jsonData, err := json.Marshal(voteData)
	if err != nil {
		return "", err
	}

	// Ensure key is 32 bytes for AES-256
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	// Create cipher block
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptVote decrypts encrypted vote data
func DecryptVote(encryptedData string, key string) (*models.VoteData, error) {
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	// Ensure key is 32 bytes
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, errors.New("encryption key must be 32 bytes")
	}

	// Create cipher block
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	var voteData models.VoteData
	if err := json.Unmarshal(plaintext, &voteData); err != nil {
		return nil, err
	}

	return &voteData, nil
}
