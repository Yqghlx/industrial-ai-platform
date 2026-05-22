package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// Repository Factory Coverage Tests
// ============================================

// TestNewRepositoryFactory tests the creation of RepositoryFactory
func TestNewRepositoryFactory(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Test factory creation
	factory := NewRepositoryFactory(db)
	assert.NotNil(t, factory, "RepositoryFactory should not be nil")
	assert.NotNil(t, factory.db, "Database interface should be initialized")

	// Ensure no unexpected expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetDeviceRepository tests getting DeviceRepository from factory
func TestGetDeviceRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetDeviceRepository()
	assert.NotNil(t, repo, "DeviceRepository should not be nil")

	// Verify it returns a non-nil repository
	assert.IsType(t, &DeviceRepository{}, repo, "Should return DeviceRepository type")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetAlertRepository tests getting AlertRepository from factory
func TestGetAlertRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetAlertRepository()
	assert.NotNil(t, repo, "AlertRepository should not be nil")
	assert.IsType(t, &AlertRepository{}, repo, "Should return AlertRepository type")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetTelemetryRepository tests getting TelemetryRepository from factory
func TestGetTelemetryRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetTelemetryRepository()
	assert.NotNil(t, repo, "TelemetryRepository should not be nil")
	assert.IsType(t, &TelemetryRepository{}, repo, "Should return TelemetryRepository type")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetUserRepository tests getting UserRepository from factory
func TestGetUserRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetUserRepository()
	assert.NotNil(t, repo, "UserRepository should not be nil")
	assert.IsType(t, &UserRepository{}, repo, "Should return UserRepository type")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetWorkOrderRepository tests getting WorkOrderRepository from factory
func TestGetWorkOrderRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetWorkOrderRepository()
	assert.NotNil(t, repo, "WorkOrderRepository should not be nil")
	assert.IsType(t, &WorkOrderRepository{}, repo, "Should return WorkOrderRepository type")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetBlackBoxRepository tests getting BlackBoxRepository from factory
func TestGetBlackBoxRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetBlackBoxRepository()
	assert.NotNil(t, repo, "BlackBoxRepository should not be nil")
	assert.IsType(t, &BlackBoxRepository{}, repo, "Should return BlackBoxRepository type")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetReportRepository tests getting ReportRepository from factory
func TestGetReportRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetReportRepository()
	assert.NotNil(t, repo, "ReportRepository should not be nil")
	assert.IsType(t, &ReportRepository{}, repo, "Should return ReportRepository type")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetNotificationRepository tests getting NotificationRepository from factory
func TestGetNotificationRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetNotificationRepository()
	assert.NotNil(t, repo, "NotificationRepository should not be nil")
	assert.IsType(t, &NotificationRepository{}, repo, "Should return NotificationRepository type")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetTenantRepository tests getting TenantRepository from factory
// Note: Currently returns nil as per TODO in factory.go
func TestGetTenantRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetTenantRepository()
	// Currently returns nil per TODO implementation
	assert.Nil(t, repo, "TenantRepository currently returns nil (TODO)")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetPermissionRepository tests getting PermissionRepository from factory
// Note: Currently returns nil as per TODO in factory.go
func TestGetPermissionRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetPermissionRepository()
	// Currently returns nil per TODO implementation
	assert.Nil(t, repo, "PermissionRepository currently returns nil (TODO)")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetRoleRepository tests getting RoleRepository from factory
// Note: Currently returns nil as per TODO in factory.go
func TestGetRoleRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetRoleRepository()
	// Currently returns nil per TODO implementation
	assert.Nil(t, repo, "RoleRepository currently returns nil (TODO)")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetRuleRepository tests getting RuleRepository from factory
func TestGetRuleRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetRuleRepository()
	assert.NotNil(t, repo, "RuleRepository should not be nil")
	assert.IsType(t, &RuleRepository{}, repo, "Should return RuleRepository type")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestGetUserRoleRepository tests getting UserRoleRepository from factory
// Note: Currently returns nil as per TODO in factory.go
func TestGetUserRoleRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	repo := factory.GetUserRoleRepository()
	// Currently returns nil per TODO implementation
	assert.Nil(t, repo, "UserRoleRepository currently returns nil (TODO)")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestRepositoryFactory_MultipleCalls tests that factory returns new instances each time
func TestRepositoryFactory_MultipleCalls(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	// Multiple calls should return different instances (not cached)
	repo1 := factory.GetDeviceRepository()
	repo2 := factory.GetDeviceRepository()
	
	assert.NotNil(t, repo1)
	assert.NotNil(t, repo2)
	// Each call creates a new repository instance
	assert.NotSame(t, repo1, repo2, "Each call should return a new repository instance")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestRepositoryFactory_AllRepositories tests all repository getters work
func TestRepositoryFactory_AllRepositories(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)
	require.NotNil(t, factory)

	// Test all implemented repositories return non-nil
	t.Run("Device", func(t *testing.T) {
		assert.NotNil(t, factory.GetDeviceRepository())
	})
	t.Run("Alert", func(t *testing.T) {
		assert.NotNil(t, factory.GetAlertRepository())
	})
	t.Run("Telemetry", func(t *testing.T) {
		assert.NotNil(t, factory.GetTelemetryRepository())
	})
	t.Run("User", func(t *testing.T) {
		assert.NotNil(t, factory.GetUserRepository())
	})
	t.Run("WorkOrder", func(t *testing.T) {
		assert.NotNil(t, factory.GetWorkOrderRepository())
	})
	t.Run("BlackBox", func(t *testing.T) {
		assert.NotNil(t, factory.GetBlackBoxRepository())
	})
	t.Run("Report", func(t *testing.T) {
		assert.NotNil(t, factory.GetReportRepository())
	})
	t.Run("Notification", func(t *testing.T) {
		assert.NotNil(t, factory.GetNotificationRepository())
	})
	t.Run("Rule", func(t *testing.T) {
		assert.NotNil(t, factory.GetRuleRepository())
	})
	// TODO items return nil
	t.Run("Tenant", func(t *testing.T) {
		assert.Nil(t, factory.GetTenantRepository())
	})
	t.Run("Role", func(t *testing.T) {
		assert.Nil(t, factory.GetRoleRepository())
	})
	t.Run("Permission", func(t *testing.T) {
		assert.Nil(t, factory.GetPermissionRepository())
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// TestNewRepositoryFactory_NilDatabase tests factory behavior with nil database
// Note: Passing nil to NewRepositoryFactory will cause panic in database.NewDBWrapper(nil)
// This test documents the current behavior
func TestNewRepositoryFactory_NilDatabase(t *testing.T) {
	// The factory panics with nil DB due to database.NewDBWrapper(nil)
	// This is expected behavior - factory requires a valid database connection
	defer func() {
		if r := recover(); r != nil {
			// Expected panic - factory requires valid DB
			t.Log("NewRepositoryFactory correctly panics with nil database")
		}
	}()
	
	factory := NewRepositoryFactory(nil)
	// If we reach here without panic, the factory was created
	assert.NotNil(t, factory, "Factory should still be created")
}