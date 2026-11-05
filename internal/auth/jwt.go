package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidToken is returned when a JWT cannot be trusted.
var ErrInvalidToken = errors.New("invalid token")

// Claims contains the identity and standard time values stored in a JWT.
type Claims struct {
	// Subject stores the admin ID as a string, following the JWT standard.
	Subject string `json:"sub"`

	// Email identifies the admin who logged in.
	Email string `json:"email"`

	// Role allows middleware to restrict routes to admins.
	Role string `json:"role"`

	// Issuer identifies the service that created the token.
	Issuer string `json:"iss"`

	// IssuedAt stores the token creation time as a Unix timestamp.
	IssuedAt int64 `json:"iat"`

	// ExpiresAt stores the token expiration time as a Unix timestamp.
	ExpiresAt int64 `json:"exp"`
}

// JWTManager creates and validates signed access tokens.
type JWTManager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

// NewJWTManager creates a JWT manager with a secret and token lifetime.
func NewJWTManager(secret string, ttl time.Duration, issuer string) (*JWTManager, error) {
	// A short secret is too easy to guess and should not be accepted.
	if len(secret) < 32 {
		return nil, errors.New("jwt secret must contain at least 32 characters")
	}

	// A non-positive lifetime would create already-expired or never-expiring tokens.
	if ttl <= 0 {
		return nil, errors.New("jwt token lifetime must be greater than zero")
	}

	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
		issuer: issuer,
	}, nil
}

// CreateAccessToken creates an HS256 JWT for an authenticated admin.
func (manager *JWTManager) CreateAccessToken(adminID uint64, email string) (string, error) {
	// Use UTC so tokens behave consistently across servers and environments.
	now := time.Now().UTC()

	// Build the JWT header and payload before signing them.
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	payload := Claims{
		Subject:   strconv.FormatUint(adminID, 10),
		Email:     email,
		Role:      "admin",
		Issuer:    manager.issuer,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(manager.ttl).Unix(),
	}

	// Encode the header and payload using base64 URL encoding without padding.
	encodedHeader, err := encodeJSON(header)
	if err != nil {
		return "", err
	}
	encodedPayload, err := encodeJSON(payload)
	if err != nil {
		return "", err
	}

	// The signature protects both the header and the payload from modification.
	unsignedToken := encodedHeader + "." + encodedPayload
	signature := manager.sign(unsignedToken)

	return unsignedToken + "." + signature, nil
}

// ValidateAccessToken verifies the signature and standard claims of a JWT.
func (manager *JWTManager) ValidateAccessToken(token string) (*Claims, error) {
	// A JWT must contain exactly three dot-separated sections.
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Only accept the HMAC SHA-256 algorithm used by this package.
	var header struct {
		Algorithm string `json:"alg"`
	}
	if err := decodeJSON(parts[0], &header); err != nil || header.Algorithm != "HS256" {
		return nil, ErrInvalidToken
	}

	// Compare signatures in constant time to reduce timing-attack information.
	expectedSignature := manager.sign(parts[0] + "." + parts[1])
	if subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(parts[2])) != 1 {
		return nil, ErrInvalidToken
	}

	// Decode the claims only after the signature has been verified.
	var claims Claims
	if err := decodeJSON(parts[1], &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Check the issuer when one was configured for this service.
	if manager.issuer != "" && claims.Issuer != manager.issuer {
		return nil, ErrInvalidToken
	}

	// Reject tokens that have expired or contain invalid time values.
	now := time.Now().Unix()
	if claims.ExpiresAt <= 0 || now >= claims.ExpiresAt || claims.IssuedAt <= 0 || claims.IssuedAt > now {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}

// sign creates the base64 URL encoded HS256 signature for an unsigned token.
func (manager *JWTManager) sign(unsignedToken string) string {
	// HMAC authenticates the token using the private server-side secret.
	hash := hmac.New(sha256.New, manager.secret)
	_, _ = hash.Write([]byte(unsignedToken))

	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}

// encodeJSON serializes a value and applies JWT-compatible base64 encoding.
func encodeJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode jwt data: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

// decodeJSON decodes JWT base64 data into the provided destination.
func decodeJSON(encodedValue string, destination any) error {
	data, err := base64.RawURLEncoding.DecodeString(encodedValue)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, destination)
}
