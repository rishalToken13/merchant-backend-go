package auth

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserUID string `json:"uid"`
	Email   string `json:"email,omitempty"`
	Role    string `json:"role"`

	// Canonical merchant_id (bytes32) encoded as 0x-prefixed hex string for JWT/clients
	MerchantID string `json:"merchant_id,omitempty"`

	jwt.RegisteredClaims
}

type JWTManager struct {
	Secret []byte
	Issuer string
	TTL    time.Duration
}

func NewJWTManager(secret, issuer string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		Secret: []byte(secret),
		Issuer: issuer,
		TTL:    ttl,
	}
}

// MerchantIDBytesToHex converts a 32-byte merchant_id to "0x..." hex string.
// If nil/empty, returns "".
func MerchantIDBytesToHex(merchantID []byte) (string, error) {
	if len(merchantID) == 0 {
		return "", nil
	}
	if len(merchantID) != 32 {
		return "", fmt.Errorf("merchant_id must be 32 bytes, got %d", len(merchantID))
	}
	return "0x" + hex.EncodeToString(merchantID), nil
}

// MerchantIDHexToBytes converts "0x..." hex merchant_id into 32 bytes.
func MerchantIDHexToBytes(s string) ([]byte, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return nil, nil
	}
	s = strings.TrimPrefix(s, "0x")
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decode merchant_id: %w", err)
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("merchant_id must decode to 32 bytes, got %d", len(b))
	}
	return b, nil
}

// Sign creates a JWT. merchantID is raw 32 bytes from DB (BYTEA).
// It is encoded as hex string in the token for portability.
func (m *JWTManager) Sign(userUID, email, role string, merchantID []byte) (token string, expiresAt time.Time, err error) {
	now := time.Now()
	exp := now.Add(m.TTL)

	merchantHex, err := MerchantIDBytesToHex(merchantID)
	if err != nil {
		return "", time.Time{}, err
	}

	claims := Claims{
		UserUID:    userUID,
		Email:      email,
		Role:       role,
		MerchantID: merchantHex,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(m.Secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

func (m *JWTManager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return m.Secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt parse: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
