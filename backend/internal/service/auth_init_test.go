package service

import (
	"github.com/stretchr/testify/mock"
)

// MockJWTService implements JWTServiceInterface for testing
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateAccessToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	args := m.Called(userID, username, role, tenantID, tokenVersion)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockJWTService) GenerateRefreshToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	args := m.Called(userID, username, role, tenantID, tokenVersion)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockJWTService) GenerateTokenPair(userID int, username, role, tenantID string, tokenVersion int) (*TokenPair, error) {
	args := m.Called(userID, username, role, tenantID, tokenVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenPair), args.Error(1)
}

func (m *MockJWTService) ParseToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *MockJWTService) RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	args := m.Called(refreshTokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenPair), args.Error(1)
}

func (m *MockJWTService) RevokeToken(tokenString string) error {
	args := m.Called(tokenString)
	return args.Error(0)
}

func (m *MockJWTService) RevokeAllUserTokens(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *MockJWTService) GetSecret() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}
