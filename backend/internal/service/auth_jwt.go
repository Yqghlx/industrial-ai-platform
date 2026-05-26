package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// JWTInitError JWT 初始化错误
type JWTInitError struct {
	Message string
}

func (e *JWTInitError) Error() string {
	return e.Message
}

// Claims represents JWT claims
type Claims struct {
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	TenantID     string `json:"tenant_id"`
	TokenType    string `json:"token_type"`
	TokenID      string `json:"token_id"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// JWTService handles JWT operations
type JWTService struct {
	secret         []byte
	tokenBlacklist TokenBlacklistInterface
	userTokenStore UserTokenStoreInterface
}

func NewJWTService(secret string) (*JWTService, error) {
	if secret == "" {
		return nil, &JWTInitError{Message: "JWT_SECRET is required"}
	}
	if len(secret) < MinSecretLength {
		return nil, &JWTInitError{
			Message: fmt.Sprintf("JWT_SECRET must be at least %d characters", MinSecretLength),
		}
	}
	return &JWTService{secret: []byte(secret)}, nil
}

func (s *JWTService) SetTokenBlacklist(blacklist TokenBlacklistInterface) {
	s.tokenBlacklist = blacklist
}

func (s *JWTService) SetUserTokenStore(store UserTokenStoreInterface) {
	s.userTokenStore = store
}

func (s *JWTService) GetSecret() []byte {
	return s.secret
}

func (s *JWTService) GenerateAccessToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	tokenID := generateTokenID()
	claims := Claims{
		UserID:       userID,
		Username:     username,
		Role:         role,
		TenantID:     tenantID,
		TokenType:    "access",
		TokenID:      tokenID,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    TokenIssuer,
			Subject:   fmt.Sprintf("user:%d", userID),
			ID:        tokenID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	return tokenString, tokenID, err
}

func (s *JWTService) GenerateRefreshToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	tokenID := generateTokenID()
	claims := Claims{
		UserID:       userID,
		Username:     username,
		Role:         role,
		TenantID:     tenantID,
		TokenType:    "refresh",
		TokenID:      tokenID,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    TokenIssuer,
			Subject:   fmt.Sprintf("user:%d", userID),
			ID:        tokenID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	return tokenString, tokenID, err
}

func (s *JWTService) GenerateTokenPair(userID int, username, role, tenantID string, tokenVersion int) (*TokenPair, error) {
	accessToken, _, err := s.GenerateAccessToken(userID, username, role, tenantID, tokenVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	refreshToken, _, err := s.GenerateRefreshToken(userID, username, role, tenantID, tokenVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		ctx := context.Background()
		if s.tokenBlacklist != nil && s.tokenBlacklist.Exists(ctx, claims.TokenID) {
			return nil, errors.New("token has been revoked")
		}
		if s.tokenBlacklist != nil {
			revokedAt, err := s.tokenBlacklist.GetUserRevocation(ctx, claims.UserID)
			if err == nil && !revokedAt.IsZero() {
				if claims.IssuedAt != nil && claims.IssuedAt.Before(revokedAt) {
					return nil, errors.New("token has been revoked (user-level)")
				}
			}
		}
		if claims.Issuer != TokenIssuer {
			return nil, errors.New("invalid token issuer")
		}
		if s.userTokenStore != nil {
			currentVersion, err := s.userTokenStore.GetTokenVersion(ctx, claims.UserID)
			if err != nil {
				logger.L().Warn("failed to get token version", zap.Error(err))
			} else if claims.TokenVersion != currentVersion {
				return nil, errors.New("token has been revoked (version mismatch)")
			}
		}
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func (s *JWTService) RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := s.ParseToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	if claims.TokenType != "refresh" {
		return nil, errors.New("token is not a refresh token")
	}
	tokenVersion := claims.TokenVersion
	if s.userTokenStore != nil {
		currentVersion, err := s.userTokenStore.GetTokenVersion(context.Background(), claims.UserID)
		if err != nil {
			logger.L().Warn("failed to get token version", zap.Error(err))
		} else {
			tokenVersion = currentVersion
		}
	}
	return s.GenerateTokenPair(claims.UserID, claims.Username, claims.Role, claims.TenantID, tokenVersion)
}

func (s *JWTService) RevokeToken(tokenString string) error {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return err
	}
	if s.tokenBlacklist == nil {
		return errors.New("token blacklist not initialized")
	}
	remainingTime := time.Until(claims.ExpiresAt.Time)
	if remainingTime <= 0 {
		return nil
	}
	return s.tokenBlacklist.Add(context.Background(), claims.TokenID, remainingTime)
}

func (s *JWTService) RevokeAllUserTokens(userID int) error {
	ctx := context.Background()
	if s.tokenBlacklist != nil {
		now := time.Now()
		err := s.tokenBlacklist.AddUserRevocation(ctx, userID, now, RefreshTokenDuration)
		if err != nil {
			logger.L().Warn("failed to add user revocation", zap.Error(err))
		}
	}
	if s.userTokenStore == nil {
		return errors.New("user token store not initialized")
	}
	return s.userTokenStore.UpdateTokenVersion(ctx, userID)
}

func generateTokenID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
	}
	return hex.EncodeToString(bytes)
}