package jwt_test

import (
	"testing"

	"bangungrid/services/auth-service/internal/adapters/security/jwt"
	"bangungrid/services/auth-service/internal/core/domain"
)

func TestRSAJWTService_GenerateAndValidate(t *testing.T) {
	jwtSvc, err := jwt.NewRSAJWTService(24, 7)
	if err != nil {
		t.Fatalf("failed to create RSA JWT service: %v", err)
	}

	claims := domain.TokenClaims{
		UserID:         "usr-12345",
		TenantID:       "tnt-001",
		OrganizationID: "org-b2b-001",
		UserType:       domain.UserTypeB2BContractor,
		Role:           "CONTRACTOR_ADMIN",
		Email:          "kontraktor@adhimix.co.id",
	}

	tokens, err := jwtSvc.GenerateTokens(claims)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Fatalf("expected non-empty refresh token")
	}
	if tokens.TokenType != "Bearer" {
		t.Fatalf("expected Bearer token type, got %s", tokens.TokenType)
	}

	// Validate refresh token
	parsedClaims, err := jwtSvc.ValidateRefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected valid refresh token, got error: %v", err)
	}

	if parsedClaims.UserID != claims.UserID {
		t.Errorf("expected UserID %s, got %s", claims.UserID, parsedClaims.UserID)
	}
	if parsedClaims.TenantID != claims.TenantID {
		t.Errorf("expected TenantID %s, got %s", claims.TenantID, parsedClaims.TenantID)
	}
	if parsedClaims.Role != claims.Role {
		t.Errorf("expected Role %s, got %s", claims.Role, parsedClaims.Role)
	}
}

func TestRSAJWTService_GetJWKS(t *testing.T) {
	jwtSvc, err := jwt.NewRSAJWTService(24, 7)
	if err != nil {
		t.Fatalf("failed to create RSA JWT service: %v", err)
	}

	jwks, err := jwtSvc.GetJWKS()
	if err != nil {
		t.Fatalf("failed to get JWKS: %v", err)
	}

	keysRaw, ok := jwks["keys"]
	if !ok {
		t.Fatalf("expected 'keys' field in JWKS")
	}

	keysList, ok := keysRaw.([]any)
	if !ok || len(keysList) == 0 {
		t.Fatalf("expected non-empty keys array in JWKS")
	}

	firstKey, ok := keysList[0].(map[string]any)
	if !ok {
		t.Fatalf("expected key object in JWKS")
	}

	if firstKey["kty"] != "RSA" {
		t.Errorf("expected kty=RSA, got %v", firstKey["kty"])
	}
	if firstKey["alg"] != "RS256" {
		t.Errorf("expected alg=RS256, got %v", firstKey["alg"])
	}
	if firstKey["use"] != "sig" {
		t.Errorf("expected use=sig, got %v", firstKey["use"])
	}
	if firstKey["n"] == "" || firstKey["e"] == "" {
		t.Errorf("expected non-empty modulus and exponent")
	}
}
