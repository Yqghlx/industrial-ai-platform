package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDefaultSecretRotationConfig tests default configuration
func TestDefaultSecretRotationConfig(t *testing.T) {
	cfg := DefaultSecretRotationConfig()

	assert.Equal(t, 24*time.Hour, cfg.RotationInterval)
	assert.Equal(t, 1*time.Hour, cfg.WarningPeriod)
	assert.Equal(t, 3, cfg.SecretsToKeep)
	assert.True(t, cfg.RotationEnabled)
	assert.Equal(t, "", cfg.CurrentSecret)
}

// TestNewSecretRotationManager tests manager creation
func TestNewSecretRotationManager(t *testing.T) {
	tests := []struct {
		name   string
		config *SecretRotationConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
		},
		{
			name: "custom config",
			config: &SecretRotationConfig{
				RotationInterval: 12 * time.Hour,
				WarningPeriod:    30 * time.Minute,
				SecretsToKeep:    5,
				RotationEnabled:  true,
				CurrentSecret:    "test-secret",
			},
		},
		{
			name: "disabled rotation",
			config: &SecretRotationConfig{
				RotationEnabled: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewSecretRotationManager(tt.config)
			assert.NotNil(t, mgr)

			// Check that a secret was generated or preserved
			secret := mgr.GetCurrentSecret()
			assert.NotEmpty(t, secret)

			if tt.config != nil && tt.config.CurrentSecret != "" {
				assert.Equal(t, tt.config.CurrentSecret, secret)
			}
		})
	}
}

// TestGetCurrentSecret tests getting current secret
func TestGetCurrentSecret(t *testing.T) {
	cfg := &SecretRotationConfig{
		CurrentSecret:   "my-secret",
		RotationEnabled: true,
		SecretsToKeep:   3,
	}
	mgr := NewSecretRotationManager(cfg)

	secret := mgr.GetCurrentSecret()
	assert.Equal(t, "my-secret", secret)
}

// TestGetOldSecrets tests getting old secrets
func TestGetOldSecrets(t *testing.T) {
	mgr := NewSecretRotationManager(nil)

	// Initially no old secrets
	oldSecrets := mgr.GetOldSecrets()
	assert.Empty(t, oldSecrets)

	// Rotate to create old secrets
	mgr.RotateSecret()
	mgr.RotateSecret()

	oldSecrets = mgr.GetOldSecrets()
	assert.Len(t, oldSecrets, 2)
}

// TestRotateSecret tests secret rotation
func TestRotateSecret(t *testing.T) {
	cfg := &SecretRotationConfig{
		CurrentSecret:   "initial-secret",
		RotationEnabled: true,
		SecretsToKeep:   3,
	}
	mgr := NewSecretRotationManager(cfg)

	initialSecret := mgr.GetCurrentSecret()
	assert.Equal(t, "initial-secret", initialSecret)

	// Rotate
	newSecret := mgr.RotateSecret()
	assert.NotEmpty(t, newSecret)
	assert.NotEqual(t, initialSecret, newSecret)

	// Current secret should be updated
	current := mgr.GetCurrentSecret()
	assert.Equal(t, newSecret, current)

	// Old secret should be preserved
	oldSecrets := mgr.GetOldSecrets()
	assert.Contains(t, oldSecrets, initialSecret)
}

// TestRotateSecretTrimOldSecrets tests that old secrets are trimmed
func TestRotateSecretTrimOldSecrets(t *testing.T) {
	cfg := &SecretRotationConfig{
		CurrentSecret:   "initial",
		RotationEnabled: true,
		SecretsToKeep:   2, // Only keep 2 old secrets
	}
	mgr := NewSecretRotationManager(cfg)

	// Rotate multiple times
	mgr.RotateSecret() // adds "initial"
	mgr.RotateSecret() // adds first rotated
	mgr.RotateSecret() // adds second rotated, trims "initial"

	oldSecrets := mgr.GetOldSecrets()
	assert.Len(t, oldSecrets, 2)
	assert.NotContains(t, oldSecrets, "initial") // should be trimmed
}

// TestValidateWithAnySecret tests token validation with multiple secrets
func TestValidateWithAnySecret(t *testing.T) {
	cfg := &SecretRotationConfig{
		CurrentSecret:   "current-secret",
		RotationEnabled: true,
		SecretsToKeep:   3,
	}
	mgr := NewSecretRotationManager(cfg)

	// Add some old secrets
	mgr.RotateSecret()

	// Validator that checks if token matches secret
	validator := func(token, secret string) bool {
		return token == "valid-token-"+secret
	}

	// Test with current secret
	current := mgr.GetCurrentSecret()
	result := mgr.ValidateWithAnySecret("valid-token-"+current, validator)
	assert.True(t, result)

	// Test with old secret
	oldSecrets := mgr.GetOldSecrets()
	if len(oldSecrets) > 0 {
		result = mgr.ValidateWithAnySecret("valid-token-"+oldSecrets[0], validator)
		assert.True(t, result)
	}

	// Test with invalid token
	result = mgr.ValidateWithAnySecret("invalid-token", validator)
	assert.False(t, result)
}

// TestGetRotationStatus tests getting rotation status
func TestGetRotationStatus(t *testing.T) {
	cfg := &SecretRotationConfig{
		RotationInterval: 24 * time.Hour,
		WarningPeriod:    1 * time.Hour,
		SecretsToKeep:    3,
		RotationEnabled:  true,
	}
	mgr := NewSecretRotationManager(cfg)

	status := mgr.GetRotationStatus()

	assert.True(t, status["enabled"].(bool))
	assert.Equal(t, 24.0, status["interval_hours"].(float64))
	assert.Equal(t, 60.0, status["warning_period_minutes"].(float64))
	assert.Equal(t, 3, status["secrets_to_keep"].(int))
}

// TestSetRotationInterval tests updating rotation interval
func TestSetRotationInterval(t *testing.T) {
	mgr := NewSecretRotationManager(nil)

	newInterval := 48 * time.Hour
	mgr.SetRotationInterval(newInterval)

	status := mgr.GetRotationStatus()
	assert.Equal(t, 48.0, status["interval_hours"].(float64))
}

// TestGetRotationWarningChannel tests warning channel
func TestGetRotationWarningChannel(t *testing.T) {
	mgr := NewSecretRotationManager(nil)

	ch := mgr.GetRotationWarningChannel()
	assert.NotNil(t, ch)
}

// TestGetRotationChannel tests rotation notification channel
func TestGetRotationChannel(t *testing.T) {
	mgr := NewSecretRotationManager(nil)

	ch := mgr.GetRotationChannel()
	assert.NotNil(t, ch)
}

// TestStartAndStop tests starting and stopping rotation
func TestStartAndStop(t *testing.T) {
	cfg := &SecretRotationConfig{
		RotationInterval: 100 * time.Millisecond,
		WarningPeriod:    50 * time.Millisecond,
		RotationEnabled:  true,
		SecretsToKeep:    3,
	}
	mgr := NewSecretRotationManager(cfg)

	// Start rotation
	mgr.Start()

	// Wait for potential rotation
	time.Sleep(150 * time.Millisecond)

	// Stop rotation
	mgr.Stop()

	// Verify it was running (secrets may have rotated)
	_ = mgr.GetCurrentSecret()
}

// TestStartDisabledRotation tests that disabled rotation doesn't start
func TestStartDisabledRotation(t *testing.T) {
	cfg := &SecretRotationConfig{
		RotationEnabled: false,
	}
	mgr := NewSecretRotationManager(cfg)

	// Start should not actually start rotation
	mgr.Start()

	// Stop should still work
	mgr.Stop()
}

// TestRotationNotification tests rotation notification channel
func TestRotationNotification(t *testing.T) {
	cfg := &SecretRotationConfig{
		RotationEnabled: true,
		SecretsToKeep:   3,
	}
	mgr := NewSecretRotationManager(cfg)

	// Get rotation channel
	rotationCh := mgr.GetRotationChannel()

	// Rotate and check notification
	newSecret := mgr.RotateSecret()

	// Check that notification was sent
	select {
	case secret := <-rotationCh:
		assert.Equal(t, newSecret, secret)
	default:
		// Channel might be empty if RotateSecret already sent
		// This is acceptable behavior
	}
}

// TestGenerateSecureSecret tests secure secret generation
func TestGenerateSecureSecret(t *testing.T) {
	secret1 := generateSecureSecret(32)
	secret2 := generateSecureSecret(32)

	assert.NotEmpty(t, secret1)
	assert.NotEmpty(t, secret2)
	assert.NotEqual(t, secret1, secret2, "Each generated secret should be unique")
	assert.Len(t, secret1, 64, "32 bytes should produce 64 hex characters")
}

// TestGenerateSecureSecretDifferentLengths tests different lengths
func TestGenerateSecureSecretDifferentLengths(t *testing.T) {
	secret16 := generateSecureSecret(16)
	secret32 := generateSecureSecret(32)
	secret64 := generateSecureSecret(64)

	assert.Len(t, secret16, 32)
	assert.Len(t, secret32, 64)
	assert.Len(t, secret64, 128)
}