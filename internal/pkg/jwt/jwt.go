package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrorInvalidToken = errors.New("invalid token")

const minSecretBytes = 32

type Signer struct {
	secret []byte
	ttl    time.Duration
}

func NewSigner(secret string, ttl time.Duration) (*Signer, error) {
	if err := validateSecret(secret); err != nil {
		return nil, err
	}
	return &Signer{secret: []byte(secret), ttl: ttl}, nil
}

func validateSecret(secret string) error {
	if len(secret) < minSecretBytes {
		return fmt.Errorf("jwt: secret must be at least %d bytes (got %d)", minSecretBytes, len(secret))
	}
	for _, b := range []byte(secret) {
		if b <= ' ' || b == 0x7f {
			return errors.New("jwt: secret must not contain whitespace or control characters")
		}
	}
	return nil
}

type claims struct {
	jwt.RegisteredClaims
}

func (s *Signer) Sign(userID int64) (string, error) {
	now := time.Now()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	})
	out, err := tok.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return out, nil
}

func (s *Signer) Parse(token string) (int64, error) {
	var c claims

	tok, err := jwt.ParseWithClaims(token, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return 0, ErrorInvalidToken
	}
	if !tok.Valid {
		return 0, ErrorInvalidToken
	}
	userID, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return 0, ErrorInvalidToken
	}
	return userID, nil
}
