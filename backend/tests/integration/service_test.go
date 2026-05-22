package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/database"

	"github.com/industrial-ai/platform/internal/service"
)

// ============================================
// UserService Integration Tests
// ============================================

func TestUserService_Integration_Create(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(database.NewDBWrapper(testDB))
	_ = service.NewUserService(userRepo)

	username := "integration_user"
	passwordHash := "hashed_password"

	_, err := testDB.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, email, role, tenant_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		username, passwordHash, "integration@test.com", "user", "1", time.Now(), time.Now(),
	)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE username = $1",
		username,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestUserService_Integration_GetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(database.NewDBWrapper(testDB))
	userService := service.NewUserService(userRepo)

	var userID int
	err := testDB.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash, email, role, tenant_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"test_get_user", "hash", "get@test.com", "user", "1", time.Now(), time.Now(),
	).Scan(&userID)
	require.NoError(t, err)

	user, err := userService.GetByID(userID)
	require.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "test_get_user", user.Username)
}

func TestUserService_Integration_GetTokenVersion(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(database.NewDBWrapper(testDB))
	userService := service.NewUserService(userRepo)

	var userID int
	err := testDB.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash, email, token_version, tenant_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"token_user", "hash", "token@test.com", 5, "1", time.Now(), time.Now(),
	).Scan(&userID)
	require.NoError(t, err)

	version, err := userService.GetTokenVersion(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 5, version)
}

func TestUserService_Integration_UpdateTokenVersion(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(database.NewDBWrapper(testDB))
	userService := service.NewUserService(userRepo)

	var userID int
	err := testDB.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash, email, token_version, tenant_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"version_user", "hash", "version@test.com", 1, "1", time.Now(), time.Now(),
	).Scan(&userID)
	require.NoError(t, err)

	err = userService.UpdateTokenVersion(ctx, userID)
	require.NoError(t, err)

	var newVersion int
	err = testDB.QueryRowContext(ctx,
		"SELECT token_version FROM users WHERE id = $1",
		userID,
	).Scan(&newVersion)
	require.NoError(t, err)
	assert.Equal(t, 2, newVersion)
}

func TestUserService_Integration_UpdatePassword(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(database.NewDBWrapper(testDB))
	userService := service.NewUserService(userRepo)

	var userID int
	err := testDB.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash, email, tenant_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"pwd_user", "old_hash", "pwd@test.com", "1", time.Now(), time.Now(),
	).Scan(&userID)
	require.NoError(t, err)

	newHash := "new_hashed_password"
	err = userService.UpdatePassword(userID, newHash)
	require.NoError(t, err)

	var currentHash string
	err = testDB.QueryRowContext(ctx,
		"SELECT password_hash FROM users WHERE id = $1",
		userID,
	).Scan(&currentHash)
	require.NoError(t, err)
	assert.Equal(t, newHash, currentHash)
}

// ============================================
// DeviceService Integration Tests
// ============================================

func TestDeviceService_Integration_Create(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(testDB))
	deviceService := service.NewDeviceService(deviceRepo, nil)

	device := &model.Device{
		ID:          "device-001",
		Name:        "Test Device",
		Type:        "sensor",
		Status:      "active",
		Description: "Integration test device",
		Location:    "test-location",
	}

	err := deviceService.Create(ctx, device)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM devices WHERE id = $1",
		device.ID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDeviceService_Integration_GetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(testDB))
	deviceService := service.NewDeviceService(deviceRepo, nil)

	_, err := testDB.ExecContext(ctx,
		`INSERT INTO devices (id, name, type, status, location, description, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		"device-get-001", "Get Device", "sensor", "active", "location-1", "test", time.Now(), time.Now(),
	)
	require.NoError(t, err)

	device, err := deviceService.GetByID(ctx, "device-get-001")
	require.NoError(t, err)
	assert.Equal(t, "device-get-001", device.ID)
	assert.Equal(t, "Get Device", device.Name)
}

func TestDeviceService_Integration_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(testDB))
	deviceService := service.NewDeviceService(deviceRepo, nil)

	_, err := testDB.ExecContext(ctx,
		`INSERT INTO devices (id, name, type, status, location, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		"device-del-001", "Delete Device", "sensor", "active", "loc-2", time.Now(), time.Now(),
	)
	require.NoError(t, err)

	err = deviceService.Delete(ctx, "device-del-001")
	require.NoError(t, err)

	var count int
	err = testDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM devices WHERE id = $1",
		"device-del-001",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeviceService_Integration_List(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()
	deviceRepo := repository.NewDeviceRepository(database.NewDBWrapper(testDB))
	deviceService := service.NewDeviceService(deviceRepo, nil)

	for i := 1; i <= 5; i++ {
		_, err := testDB.ExecContext(ctx,
			`INSERT INTO devices (id, name, type, status, location, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			fmt.Sprintf("device-list-%d", i),
			fmt.Sprintf("List Device %d", i),
			"sensor",
			"active",
			fmt.Sprintf("loc-%d", i),
			time.Now(),
			time.Now(),
		)
		require.NoError(t, err)
	}

	devices, total, err := deviceService.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, devices, 5)
}

// ============================================
// AlertService Integration Tests
// ============================================

func TestAlertService_Integration_CreateAlert(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()

	_, err := testDB.ExecContext(ctx,
		"INSERT INTO devices (id, name, type, location, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		"alert-device", "Alert Device", "sensor", "loc", time.Now(), time.Now(),
	)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx,
		`INSERT INTO alerts (device_id, message, severity, status, triggered_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		"alert-device", "Integration test alert", "high", "active", time.Now(),
	)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM alerts WHERE device_id = $1",
		"alert-device",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestAlertService_Integration_CreateRule(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()

	var ruleID int
	err := testDB.QueryRowContext(ctx,
		`INSERT INTO alert_rules (name, metric, operator, threshold, severity, device_type, enabled, cooldown_sec)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		"Integration Rule", "temperature", ">", 80.0, "high", "*", true, 300,
	).Scan(&ruleID)
	require.NoError(t, err)
	assert.Greater(t, ruleID, 0)
}

// ============================================
// Telemetry Integration Tests
// ============================================

func TestTelemetry_Integration_Insert(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()

	_, err := testDB.ExecContext(ctx,
		"INSERT INTO devices (id, name, type, location, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		"telemetry-device", "Telemetry Device", "sensor", "loc", time.Now(), time.Now(),
	)
	require.NoError(t, err)

	for i := 1; i <= 10; i++ {
		_, err := testDB.ExecContext(ctx,
			"INSERT INTO telemetry (device_id, metric, value, timestamp) VALUES ($1, $2, $3, $4)",
			"telemetry-device", "temperature", 25.0+float64(i), time.Now(),
		)
		require.NoError(t, err)
	}

	var count int
	err = testDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM telemetry WHERE device_id = $1",
		"telemetry-device",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 10, count)
}

// ============================================
// RBAC Integration Tests
// ============================================

func TestRBAC_Integration_CreateRole(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()

	var roleID int
	err := testDB.QueryRowContext(ctx,
		"INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id",
		"test_role", "Integration test role",
	).Scan(&roleID)
	require.NoError(t, err)
	assert.Greater(t, roleID, 0)

	var count int
	err = testDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM roles WHERE name = $1",
		"test_role",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRBAC_Integration_AssignRole(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()

	var userID int
	err := testDB.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash, email, tenant_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"rbac_user", "hash", "rbac@test.com", "1", time.Now(), time.Now(),
	).Scan(&userID)
	require.NoError(t, err)

	var roleID int
	err = testDB.QueryRowContext(ctx,
		"INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id",
		"assigned_role", "Role to be assigned",
	).Scan(&roleID)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx,
		"INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)",
		userID, roleID,
	)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND role_id = $2",
		userID, roleID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRBAC_Integration_CreatePermission(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()

	var permID int
	err := testDB.QueryRowContext(ctx,
		"INSERT INTO permissions (name, resource, action) VALUES ($1, $2, $3) RETURNING id",
		"read_devices", "devices", "read",
	).Scan(&permID)
	require.NoError(t, err)
	assert.Greater(t, permID, 0)
}

func TestRBAC_Integration_AssignPermission(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration tests skipped: database not available")
	}
	truncateAllTables(t)
	ctx := context.Background()

	var roleID int
	err := testDB.QueryRowContext(ctx,
		"INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id",
		"perm_role", "Role with permissions",
	).Scan(&roleID)
	require.NoError(t, err)

	var permID int
	err = testDB.QueryRowContext(ctx,
		"INSERT INTO permissions (name, resource, action) VALUES ($1, $2, $3) RETURNING id",
		"write_devices", "devices", "write",
	).Scan(&permID)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx,
		"INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)",
		roleID, permID,
	)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM role_permissions WHERE role_id = $1 AND permission_id = $2",
		roleID, permID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
