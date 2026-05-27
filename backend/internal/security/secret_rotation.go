package security

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"
)

// SEC-MED-03: Automatic JWT Secret Rotation Implementation
// This module provides automatic JWT secret rotation with configurable intervals.

// SecretRotationConfig holds configuration for secret rotation
type SecretRotationConfig struct {
	// RotationInterval is the interval between rotations (default: 24 hours)
	RotationInterval time.Duration
	// WarningPeriod is the time before rotation to send warnings (default: 1 hour)
	WarningPeriod time.Duration
	// SecretsToKeep is the number of old secrets to keep for validation (default: 3)
	SecretsToKeep int
	// CurrentSecret is the current active secret
	CurrentSecret string
	// RotationEnabled controls whether automatic rotation is active
	RotationEnabled bool
}

// DefaultSecretRotationConfig returns default configuration
func DefaultSecretRotationConfig() *SecretRotationConfig {
	return &SecretRotationConfig{
		RotationInterval: 24 * time.Hour,
		WarningPeriod:    1 * time.Hour,
		SecretsToKeep:    3,
		RotationEnabled:  true,
		CurrentSecret:    "",
	}
}

// SecretRotationManager manages JWT secret rotation
type SecretRotationManager struct {
	config        *SecretRotationConfig
	currentSecret string
	oldSecrets    []string // Keep old secrets for token validation during transition
	mu            sync.RWMutex
	stopChan      chan struct{}
	warningChan   chan time.Time // Channel to send rotation warnings
	rotationChan  chan string    // Channel to notify of new secrets
}

// NewSecretRotationManager creates a new secret rotation manager
func NewSecretRotationManager(config *SecretRotationConfig) *SecretRotationManager {
	if config == nil {
		config = DefaultSecretRotationConfig()
	}

	mgr := &SecretRotationManager{
		config:        config,
		currentSecret: config.CurrentSecret,
		oldSecrets:    make([]string, 0, config.SecretsToKeep),
		stopChan:      make(chan struct{}),
		warningChan:   make(chan time.Time, 1),
		rotationChan:  make(chan string, 1),
	}

	// Generate initial secret if not provided
	if mgr.currentSecret == "" {
		mgr.currentSecret = generateSecureSecret(32)
	}

	return mgr
}

// Start begins the automatic rotation process
func (m *SecretRotationManager) Start() {
	if !m.config.RotationEnabled {
		log.Println("[SecretRotation] Automatic rotation is disabled")
		return
	}

	go m.rotationLoop()
	log.Printf("[SecretRotation] Started with interval: %v", m.config.RotationInterval)
}

// Stop stops the automatic rotation process
func (m *SecretRotationManager) Stop() {
	close(m.stopChan)
	log.Println("[SecretRotation] Stopped")
}

// GetCurrentSecret returns the current active secret
func (m *SecretRotationManager) GetCurrentSecret() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentSecret
}

// GetOldSecrets returns old secrets for token validation
func (m *SecretRotationManager) GetOldSecrets() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.oldSecrets
}

// ValidateWithAnySecret validates a token using current or old secrets
func (m *SecretRotationManager) ValidateWithAnySecret(token string, validator func(token, secret string) bool) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try current secret first
	if validator(token, m.currentSecret) {
		return true
	}

	// Try old secrets
	for _, secret := range m.oldSecrets {
		if validator(token, secret) {
			return true
		}
	}

	return false
}

// RotateSecret manually rotates the secret
func (m *SecretRotationManager) RotateSecret() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	newSecret := generateSecureSecret(32)

	// Add current secret to old secrets list
	if m.currentSecret != "" {
		m.oldSecrets = append(m.oldSecrets, m.currentSecret)
		// Trim old secrets to configured limit
		if len(m.oldSecrets) > m.config.SecretsToKeep {
			m.oldSecrets = m.oldSecrets[len(m.oldSecrets)-m.config.SecretsToKeep:]
		}
	}

	m.currentSecret = newSecret

	// Notify rotation
	select {
	case m.rotationChan <- newSecret:
	default:
	}

	log.Printf("[SecretRotation] Secret rotated successfully. Old secrets count: %d", len(m.oldSecrets))
	return newSecret
}

// GetRotationWarningChannel returns the channel for rotation warnings
func (m *SecretRotationManager) GetRotationWarningChannel() <-chan time.Time {
	return m.warningChan
}

// GetRotationChannel returns the channel for new secret notifications
func (m *SecretRotationManager) GetRotationChannel() <-chan string {
	return m.rotationChan
}

// rotationLoop handles automatic rotation
func (m *SecretRotationManager) rotationLoop() {
	ticker := time.NewTicker(m.config.RotationInterval)
	warningTicker := time.NewTicker(m.config.RotationInterval - m.config.WarningPeriod)

	defer ticker.Stop()
	defer warningTicker.Stop()

	for {
		select {
		case <-m.stopChan:
			return

		case <-warningTicker.C:
			// Send rotation warning
			nextRotation := time.Now().Add(m.config.WarningPeriod)
			select {
			case m.warningChan <- nextRotation:
				log.Printf("[SecretRotation] Warning: Secret rotation in %v", m.config.WarningPeriod)
			default:
			}

		case <-ticker.C:
			// Perform rotation
			m.RotateSecret()
		}
	}
}

// generateSecureSecret generates a cryptographically secure secret
func generateSecureSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based generation if crypto/rand fails
		log.Printf("[SecretRotation] Warning: crypto/rand failed, using fallback: %v", err)
		return hex.EncodeToString([]byte(time.Now().String())) + hex.EncodeToString(bytes)
	}
	return hex.EncodeToString(bytes)
}

// GetRotationStatus returns current rotation status
func (m *SecretRotationManager) GetRotationStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"enabled":                m.config.RotationEnabled,
		"interval_hours":         m.config.RotationInterval.Hours(),
		"warning_period_minutes": m.config.WarningPeriod.Minutes(),
		"old_secrets_count":      len(m.oldSecrets),
		"secrets_to_keep":        m.config.SecretsToKeep,
	}
}

// SetRotationInterval updates the rotation interval
func (m *SecretRotationManager) SetRotationInterval(interval time.Duration) {
	m.mu.Lock()
	m.config.RotationInterval = interval
	m.mu.Unlock()
}
