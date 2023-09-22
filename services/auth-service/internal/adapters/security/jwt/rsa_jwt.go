package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"bangungrid/services/auth-service/internal/core/domain"
	"bangungrid/services/auth-service/internal/core/ports"
)

type RSAJWTService struct {
	privateKey        *rsa.PrivateKey
	publicKey         *rsa.PublicKey
	keyID             string
	accessExpiryHours int
	refreshExpiryDays int
}

func NewRSAJWTService(accessExpiryHours, refreshExpiryDays int) (ports.TokenServicePort, error) {
	// Buat RSA Keypair 2048-bit otomatis (atau muat dari file)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat rsa keypair: %w", err)
	}

	if accessExpiryHours <= 0 {
		accessExpiryHours = 24
	}
	if refreshExpiryDays <= 0 {
		refreshExpiryDays = 7
	}

	return &RSAJWTService{
		privateKey:        privateKey,
		publicKey:         &privateKey.PublicKey,
		keyID:             "bangungrid-key-1",
		accessExpiryHours: accessExpiryHours,
		refreshExpiryDays: refreshExpiryDays,
	}, nil
}

type CustomClaims struct {
	UserID         string          `json:"sub"`
	TenantID       string          `json:"tenant_id"`
	OrganizationID string          `json:"organization_id,omitempty"`
	UserType       domain.UserType `json:"user_type"`
	Role           string          `json:"role"`
	Email          string          `json:"email"`
	jwt.RegisteredClaims
}

func (s *RSAJWTService) GenerateTokens(claims domain.TokenClaims) (*domain.AuthToken, error) {
	now := time.Now().UTC()
	accessExpiry := now.Add(time.Duration(s.accessExpiryHours) * time.Hour)
	refreshExpiry := now.Add(time.Duration(s.refreshExpiryDays) * 24 * time.Hour)

	// 1. Access Token (RS256)
	accessClaims := CustomClaims{
		UserID:         claims.UserID,
		TenantID:       claims.TenantID,
		OrganizationID: claims.OrganizationID,
		UserType:       claims.UserType,
		Role:           claims.Role,
		Email:          claims.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "bangungrid-auth-service",
			Subject:   claims.UserID,
			Audience:  jwt.ClaimStrings{"bangungrid-api"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	token.Header["kid"] = s.keyID

	accessTokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani access token: %w", err)
	}

	// 2. Refresh Token (RS256)
	refreshClaims := CustomClaims{
		UserID:   claims.UserID,
		TenantID: claims.TenantID,
		Role:     claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "bangungrid-auth-service",
			Subject:   claims.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			ID:        uuid.New().String(),
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenObj.Header["kid"] = s.keyID

	refreshTokenString, err := refreshTokenObj.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani refresh token: %w", err)
	}

	return &domain.AuthToken{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessExpiryHours * 3600),
		ExpiresAt:    accessExpiry,
	}, nil
}

func (s *RSAJWTService) ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("metode signing tidak terduga: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return &domain.TokenClaims{
		UserID:         claims.UserID,
		TenantID:       claims.TenantID,
		OrganizationID: claims.OrganizationID,
		UserType:       claims.UserType,
		Role:           claims.Role,
		Email:          claims.Email,
	}, nil
}

// GetJWKS menghasilkan format standar RFC 7517 JSON Web Key Set untuk dikonsumsi KrakenD Gateway
func (s *RSAJWTService) GetJWKS() (map[string]any, error) {
	nBytes := s.publicKey.N.Bytes()
	eBytes := big.NewInt(int64(s.publicKey.E)).Bytes()

	nBase64 := base64.RawURLEncoding.EncodeToString(nBytes)
	eBase64 := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := map[string]any{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": s.keyID,
		"n":   nBase64,
		"e":   eBase64,
	}

	return map[string]any{
		"keys": []any{jwk},
	}, nil
}
