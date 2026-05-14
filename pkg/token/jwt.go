package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

type Claims struct {
	UserID int64 `json:"user_id"`
	Exp    int64 `json:"exp"`
	Iat    int64 `json:"iat"`
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

func (m *JWTManager) GenerateToken(userID int64) (string, error) {
	now := time.Now().Unix()
	claims := Claims{
		UserID: userID,
		Exp:    now + int64(m.tokenDuration.Seconds()),
		Iat:    now,
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signatureInput := headerEncoded + "." + claimsEncoded
	h := hmac.New(sha256.New, m.secretKey)
	h.Write([]byte(signatureInput))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return signatureInput + "." + signature, nil
}

func (m *JWTManager) VerifyToken(tokenString string) (*Claims, error) {
	parts := splitString(tokenString, '.')
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signature := parts[2]

	signatureInput := headerEncoded + "." + claimsEncoded
	h := hmac.New(sha256.New, m.secretKey)
	h.Write([]byte(signatureInput))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, errors.New("invalid signature")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, errors.New("invalid claims encoding")
	}

	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, errors.New("invalid claims format")
	}

	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

func splitString(s string, sep byte) []string {
	var parts []string
	var start int
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
