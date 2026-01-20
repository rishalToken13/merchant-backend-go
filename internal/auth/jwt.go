package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserUID    string `json:"uid"`
	Role       string `json:"role"`
	MerchantID *int64  `json:"merchant_id,omitempty"`
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

func (m *JWTManager) Sign(userUID, role string, merchantID *int64) (string, error) {
	now := time.Now()

	claims := Claims{
		UserUID:    userUID,
		Role:       role,
		MerchantID: merchantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.TTL)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.Secret)
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