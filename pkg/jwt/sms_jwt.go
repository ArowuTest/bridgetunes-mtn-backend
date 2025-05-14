package jwt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	// "crypto/sha256"
	"encoding/hex"
	"fmt"

	// "io"
	"log"
	"strings"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// SMSTokenService handles JWT token generation and encryption for SMS
type SMSTokenService struct {
	accessTokenSecret  string
	jwtSecret         string
	refreshTokenSecret string
	uduxSmsEncryptKey []byte
	algorithm         string
	iv               []byte // Store the fixed IV
}

func NewSMSTokenService(cfg *config.Config) *SMSTokenService {
	// Use a fixed IV like in Node.js
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		panic(err)
	}

	// Use the raw key like Node.js, don't hash it
	key := []byte(cfg.UduxGateway.APISecret)
	// Ensure key is 32 bytes (required for AES-256)
	if len(key) < 32 {
		// Pad with zeros if shorter
		paddedKey := make([]byte, 32)
		copy(paddedKey, key)
		key = paddedKey
	} else if len(key) > 32 {
		// Truncate if longer
		key = key[:32]
	}

	return &SMSTokenService{
		accessTokenSecret:  cfg.UduxGateway.APISecret,
		jwtSecret:         cfg.UduxGateway.JWT_SECRET,
		refreshTokenSecret: cfg.UduxGateway.APISecret,
		uduxSmsEncryptKey:  key,
		algorithm:         "aes-256-cbc",
		iv:               iv,
	}
}

// SMSTokens represents the access and refresh tokens
type SMSTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// GetSMSTokens generates and encrypts SMS tokens
func (s *SMSTokenService) GetSMSTokens(userID string, userType string) (*SMSTokens, error) {
	// Create JWT payload
	payload := jwt.MapClaims{
		"sub":          userID,
		"ssId":         nil,
		"role":         []string{userType},
		"iat":          time.Now().Unix(),
		"exp":          time.Now().Add(24 * time.Hour).Unix(),
	}

	// Generate tokens
	accessToken, err := s.signSMSAccessToken(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, err := s.signSMSRefreshToken(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	log.Printf("jwt token: %+v\n", accessToken)
	// Encrypt tokens
	encryptedAccessToken := s.encryptUduxSmsTokenOriginal(accessToken)
	encryptedRefreshToken, err := s.encryptUduxSmsToken(refreshToken)


	return &SMSTokens{
		AccessToken:  encryptedAccessToken,
		RefreshToken: encryptedRefreshToken,
	}, nil
}

// signSMSAccessToken signs the access token
func (s *SMSTokenService) signSMSAccessToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// signSMSRefreshToken signs the refresh token
func (s *SMSTokenService) signSMSRefreshToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.refreshTokenSecret))
}

// encrypt encrypts the given text using AES-256-CBC
func (s *SMSTokenService) encryptUduxSmsTokenOriginal(text string) string {
	log.Printf("encryptionKey:  %+v\n", s.uduxSmsEncryptKey)
	block, err := aes.NewCipher(s.uduxSmsEncryptKey)
	if err != nil {
		panic(err)
	}

	// Create a new CBC mode encrypter
	mode := cipher.NewCBCEncrypter(block, s.iv)

	// Pad the text to be a multiple of the block size
	paddedText := pkcs7Padding([]byte(text), aes.BlockSize)

	// Encrypt the text
	encrypted := make([]byte, len(paddedText))
	mode.CryptBlocks(encrypted, paddedText)

	// Return the IV and encrypted text as a hex string


	// enc := fmt.Sprintf("%s:%s", hex.EncodeToString(iv), hex.EncodeToString(encrypted))
	enc := fmt.Sprintf("%x:%x", s.iv, encrypted)
	dec, ert := s.Decrypt(enc)
	log.Printf("dec:  %+v\n", dec)
	log.Printf("ert:  %+v\n", ert)
	return enc
}

// encryptUduxSmsToken encrypts the text using AES-256-CBC to match Node.js implementation
func (s *SMSTokenService) encryptUduxSmsToken(text string) (string, error) {
	block, err := aes.NewCipher(s.uduxSmsEncryptKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use the fixed IV
	mode := cipher.NewCBCEncrypter(block, s.iv)

	// Pad the text to be a multiple of the block size
	paddedText := pkcs7Padding([]byte(text), aes.BlockSize)

	// Encrypt the text
	encrypted := make([]byte, len(paddedText))
	mode.CryptBlocks(encrypted, paddedText)

	// Format exactly like Node.js: iv:encrypted
	return fmt.Sprintf("%s:%s", hex.EncodeToString(s.iv), hex.EncodeToString(encrypted)), nil
}

// pkcs7Padding adds PKCS#7 padding to match Node.js crypto padding
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// Decrypt decrypts the text using AES-256-CBC to match Node.js implementation
func (s *SMSTokenService) Decrypt(encryptedText string) (string, error) {
	// Split text into IV and actual ciphertext
	parts := strings.Split(encryptedText, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid encrypted text format")
	}

	// Decode hex strings
	iv, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid IV: %w", err)
	}

	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(s.uduxSmsEncryptKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	unpaddedPlaintext, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to remove padding: %w", err)
	}

	return string(unpaddedPlaintext), nil
}

// pkcs7Unpad removes PKCS#7 padding
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) {
		return nil, fmt.Errorf("invalid padding size")
	}

	// Verify padding
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}