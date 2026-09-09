package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type AccessClaims struct {
	Sub string `json:"sub"`
	Typ string `json:"typ"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
	JTI string `json:"jti"`
}

type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	if secret == "" {
		secret = "dev-only-change-me"
	}
	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (m *TokenManager) NewAccessToken(userID string) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(m.accessTTL)
	claims := AccessClaims{Sub: userID, Typ: "access", Iat: now.Unix(), Exp: exp.Unix(), JTI: randomToken(12)}
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	enc := base64.RawURLEncoding
	unsigned := enc.EncodeToString(header) + "." + enc.EncodeToString(payload)
	sig := sign(m.secret, unsigned)
	return unsigned + "." + enc.EncodeToString(sig), exp, nil
}

func (m *TokenManager) ParseAccessToken(token string) (AccessClaims, error) {
	var c AccessClaims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c, errors.New("invalid token")
	}
	enc := base64.RawURLEncoding
	sig, err := enc.DecodeString(parts[2])
	if err != nil {
		return c, errors.New("invalid token")
	}
	if !hmac.Equal(sig, sign(m.secret, parts[0]+"."+parts[1])) {
		return c, errors.New("invalid token")
	}
	payload, err := enc.DecodeString(parts[1])
	if err != nil {
		return c, errors.New("invalid token")
	}
	if err := json.Unmarshal(payload, &c); err != nil {
		return c, errors.New("invalid token")
	}
	if c.Typ != "access" || c.Sub == "" || time.Now().UTC().Unix() >= c.Exp {
		return c, errors.New("expired or invalid token")
	}
	return c, nil
}

func (m *TokenManager) NewRefreshToken() (raw, hash string, expiresAt time.Time) {
	raw = randomToken(32)
	sum := sha256.Sum256([]byte(raw))
	hash = base64.RawURLEncoding.EncodeToString(sum[:])
	return raw, hash, time.Now().UTC().Add(m.refreshTTL)
}

func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func sign(secret []byte, value string) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
