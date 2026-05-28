package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// base_repo.go - 补充测试
// ============================================

// TestUpdateTimeColumn 测试 UpdateTimeColumn 函数
func TestUpdateTimeColumn(t *testing.T) {
	now := UpdateTimeColumn()
	assert.NotNil(t, now)
	// 验证返回的时间接近当前时间
	assert.True(t, now.Before(time.Now().Add(time.Second)))
	assert.True(t, now.After(time.Now().Add(-time.Second)))
}

// TestWithTransaction_NilDB 测试 WithTransaction 在 nil db 情况下返回错误
func TestWithTransaction_NilDB(t *testing.T) {
	ctx := context.Background()
	err := WithTransaction(ctx, nil, func(tx database.TransactionInterface) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not configured for transactions")
}

// TestWithTransaction_Success 测试 WithTransaction 成功执行
func TestWithTransaction_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectCommit()

	err = WithTransaction(ctx, dbWrapper, func(tx database.TransactionInterface) error {
		return nil
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestWithTransaction_RollbackOnError 测试 WithTransaction 在错误时回滚
func TestWithTransaction_RollbackOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectRollback()

	err = WithTransaction(ctx, dbWrapper, func(tx database.TransactionInterface) error {
		return errors.New("operation failed")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestValidateTableAndColumn_ColumnValidation 测试列名验证
func TestValidateTableAndColumn_ColumnValidation(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		column  string
		wantErr bool
		errMsg  string
	}{
		{"Valid_Simple", "users", "id", false, ""},
		{"Valid_Underscore", "users", "user_name", false, ""},
		{"Valid_CamelCase", "users", "UserName", false, ""},
		{"EmptyColumn", "users", "", true, "column name cannot be empty"},
		{"InvalidStartDigit", "users", "1column", true, "must start with letter or underscore"},
		{"InvalidStartSpecial", "users", "@column", true, "must start with letter or underscore"},
		{"InvalidChar", "users", "user-name", true, "invalid character"},
		{"TooLong", "users", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, "column name too long"},
		{"InvalidTable", "invalid_table", "id", true, "invalid table name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTableAndColumn(tt.table, tt.column)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================
// permission_repo.go - WithTx 测试
// ============================================

func TestPermissionRepo_WithTx_Coverage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewPermissionRepo(dbWrapper)

	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	require.NotNil(t, txRepo)

	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPermissionRepo_GetByIDs_Coverage 测试批量获取权限
func TestPermissionRepo_GetByIDs_Coverage(t *testing.T) {
	// 空ID测试 - 这是可以测试的边界情况
	t.Run("EmptyIDs", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewPermissionRepo(database.NewDBWrapper(db))
		perms, err := repo.GetByIDs([]int{})
		assert.NoError(t, err)
		assert.Empty(t, perms)
	})

	// 成功测试 - sqlmock不支持PostgreSQL数组参数，需要真实数据库
	t.Run("Success", func(t *testing.T) {
		t.Skip("sqlmock不支持PostgreSQL ANY($1)数组参数类型")
	})

	// 查询错误测试 - sqlmock不支持PostgreSQL数组参数
	t.Run("QueryError", func(t *testing.T) {
		t.Skip("sqlmock不支持PostgreSQL ANY($1)数组参数类型")
	})
}

// TestPermissionRepo_GetByIDs_ScanError 测试扫描错误
func TestPermissionRepo_GetByIDs_ScanError(t *testing.T) {
	t.Skip("sqlmock不支持PostgreSQL ANY($1)数组参数类型")
}

// ============================================
// role_repo.go - WithTx 测试
// ============================================

func TestRoleRepo_WithTx_CoverageBoost(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewRoleRepo(dbWrapper)

	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	require.NotNil(t, txRepo)

	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ============================================
// tenant_repo.go - WithTx 测试
// ============================================

func TestTenantRepo_WithTx_Coverage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbWrapper := database.NewDBWrapper(db)
	repo := NewTenantRepo(dbWrapper)

	mock.ExpectBegin()
	tx, err := dbWrapper.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	require.NotNil(t, txRepo)

	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ============================================
// rule_repo.go (AlertRepository) - 补充测试
// ============================================

// TestAlertRepository_ListWithFilter_Coverage 测试带过滤条件的告警列表查询
func TestAlertRepository_ListWithFilter_Coverage(t *testing.T) {
	tests := []struct {
		name   string
		filter AlertFilter
		page   int
		setup  func(mock sqlmock.Sqlmock)
	}{
		{
			name:   "NoFilter",
			filter: AlertFilter{},
			page:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
				rows := sqlmock.NewRows([]string{"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at"}).
					AddRow(1, 1, "device-1", "High temp", "high", "active", time.Now(), nil)
				mock.ExpectQuery(`SELECT .* FROM alerts`).WillReturnRows(rows)
			},
		},
		{
			name:   "FilterByStatus",
			filter: AlertFilter{Status: "active"},
			page:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
				rows := sqlmock.NewRows([]string{"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at"}).
					AddRow(1, 1, "device-1", "High temp", "high", "active", time.Now(), nil)
				mock.ExpectQuery(`SELECT .* FROM alerts`).WillReturnRows(rows)
			},
		},
		{
			name:   "FilterBySeverity",
			filter: AlertFilter{Severity: "high"},
			page:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
				rows := sqlmock.NewRows([]string{"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at"}).
					AddRow(1, 1, "device-1", "High temp", "high", "active", time.Now(), nil)
				mock.ExpectQuery(`SELECT .* FROM alerts`).WillReturnRows(rows)
			},
		},
		{
			name:   "FilterByDeviceID",
			filter: AlertFilter{DeviceID: "device-1"},
			page:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
				rows := sqlmock.NewRows([]string{"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at"}).
					AddRow(1, 1, "device-1", "High temp", "high", "active", time.Now(), nil)
				mock.ExpectQuery(`SELECT .* FROM alerts`).WillReturnRows(rows)
			},
		},
		{
			name:   "FilterAll",
			filter: AlertFilter{Status: "active", Severity: "high", DeviceID: "device-1"},
			page:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				rows := sqlmock.NewRows([]string{"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at"}).
					AddRow(1, 1, "device-1", "High temp", "high", "active", time.Now(), nil)
				mock.ExpectQuery(`SELECT .* FROM alerts`).WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewAlertRepository(database.NewDBWrapper(db))
			tt.setup(mock)

			alerts, total, err := repo.ListWithFilter(context.Background(), tt.filter, tt.page, 10)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, total, 0)
			assert.NoError(t, mock.ExpectationsWereMet())
			_ = alerts
		})
	}
}

// TestAlertRepository_CountByStatus_Coverage 测试按状态计数
func TestAlertRepository_CountByStatus_Coverage(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "CountActive",
			status: "active",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts WHERE status =`).
					WithArgs("active").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
			},
			wantErr: false,
		},
		{
			name:   "CountResolved",
			status: "resolved",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts WHERE status =`).
					WithArgs("resolved").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			wantErr: false,
		},
		{
			name:   "QueryError",
			status: "active",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts WHERE status =`).
					WillReturnError(errors.New("query error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewAlertRepository(database.NewDBWrapper(db))
			tt.setup(mock)

			count, err := repo.CountByStatus(context.Background(), tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, count, 0)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAlertRepository_UpdateStatus_Coverage 测试更新告警状态
func TestAlertRepository_UpdateStatus_Coverage(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		status  string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "Success",
			id:     1,
			status: "acknowledged",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE alerts SET status =`).
					WithArgs("acknowledged", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:   "ExecError",
			id:     1,
			status: "acknowledged",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE alerts SET status =`).
					WillReturnError(errors.New("exec error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewAlertRepository(database.NewDBWrapper(db))
			tt.setup(mock)

			err = repo.UpdateStatus(context.Background(), tt.id, tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAlertRepository_GetRecentAlertsByDeviceBatch_Coverage 测试批量获取最近告警
func TestAlertRepository_GetRecentAlertsByDeviceBatch_Coverage(t *testing.T) {
	// 空规则ID测试 - 这是可以测试的边界情况
	t.Run("EmptyRuleIDs", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewAlertRepository(database.NewDBWrapper(db))

		result, err := repo.GetRecentAlertsByDeviceBatch(context.Background(), "device-1", []int{}, 60)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	// 成功测试 - sqlmock不支持PostgreSQL数组参数
	t.Run("Success", func(t *testing.T) {
		t.Skip("sqlmock不支持PostgreSQL ANY($2)数组参数类型")
	})

	// 查询错误测试 - sqlmock不支持PostgreSQL数组参数
	t.Run("QueryError", func(t *testing.T) {
		t.Skip("sqlmock不支持PostgreSQL ANY($2)数组参数类型")
	})
}

// TestAlertRepository_ArchiveOldAlerts_Coverage 测试归档旧告警
func TestAlertRepository_ArchiveOldAlerts_Coverage(t *testing.T) {
	tests := []struct {
		name    string
		daysOld int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:    "Success",
			daysOld: 30,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE alerts SET status = 'archived'`).
					WithArgs(30).
					WillReturnResult(sqlmock.NewResult(0, 5))
			},
			wantErr: false,
		},
		{
			name:    "ExecError",
			daysOld: 30,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE alerts SET status = 'archived'`).
					WillReturnError(errors.New("exec error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewAlertRepository(database.NewDBWrapper(db))
			tt.setup(mock)

			count, err := repo.ArchiveOldAlerts(context.Background(), tt.daysOld)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, count, 0)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAlertRepository_GetArchivedAlerts_Coverage 测试获取已归档告警
func TestAlertRepository_GetArchivedAlerts_Coverage(t *testing.T) {
	tests := []struct {
		name     string
		deviceID string
		page     int
		setup    func(mock sqlmock.Sqlmock)
		wantErr  bool
	}{
		{
			name:     "NoDeviceFilter",
			deviceID: "",
			page:     1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
				resolvedAt := sql.NullTime{Time: time.Now(), Valid: true}
				rows := sqlmock.NewRows([]string{"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at"}).
					AddRow(1, 1, "device-1", "Archived alert", "high", "archived", time.Now(), resolvedAt)
				mock.ExpectQuery(`SELECT .* FROM alerts WHERE status = 'archived'`).WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:     "WithDeviceFilter",
			deviceID: "device-1",
			page:     1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
				resolvedAt := sql.NullTime{Time: time.Now(), Valid: true}
				rows := sqlmock.NewRows([]string{"id", "rule_id", "device_id", "message", "severity", "status", "triggered_at", "resolved_at"}).
					AddRow(1, 1, "device-1", "Archived alert", "high", "archived", time.Now(), resolvedAt)
				mock.ExpectQuery(`SELECT .* FROM alerts WHERE status = 'archived' AND device_id`).WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:     "CountError",
			deviceID: "",
			page:     1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT`).WillReturnError(errors.New("count error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewAlertRepository(database.NewDBWrapper(db))
			tt.setup(mock)

			alerts, total, err := repo.GetArchivedAlerts(context.Background(), tt.deviceID, tt.page, 10)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, total, 0)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
			_ = alerts
		})
	}
}

// TestAlertRepository_DeleteArchivedAlerts_Coverage 测试删除已归档告警
func TestAlertRepository_DeleteArchivedAlerts_Coverage(t *testing.T) {
	tests := []struct {
		name    string
		daysOld int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:    "Success",
			daysOld: 365,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM alerts WHERE status = 'archived'`).
					WithArgs(365).
					WillReturnResult(sqlmock.NewResult(0, 10))
			},
			wantErr: false,
		},
		{
			name:    "ExecError",
			daysOld: 365,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM alerts WHERE status = 'archived'`).
					WillReturnError(errors.New("exec error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewAlertRepository(database.NewDBWrapper(db))
			tt.setup(mock)

			count, err := repo.DeleteArchivedAlerts(context.Background(), tt.daysOld)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, count, 0)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAlertRepository_GetAlertStatistics_Coverage 测试获取告警统计
func TestAlertRepository_GetAlertStatistics_Coverage(t *testing.T) {
	t.Skip("GetAlertStatistics需要真实数据库环境进行测试，sqlmock不支持复杂的多查询统计")
}

// ============================================
// telemetry_repo.go - 补充测试
// ============================================

// TestWorkOrderRepository_CountOpen_Coverage 测试计数开放工单
func TestWorkOrderRepository_CountOpen_Coverage(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "Success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM work_orders WHERE status IN`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
			},
			wantErr: false,
		},
		{
			name: "QueryError",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM work_orders WHERE status IN`).
					WillReturnError(errors.New("query error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewWorkOrderRepository(database.NewDBWrapper(db))
			tt.setup(mock)

			count, err := repo.CountOpen(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, count, 0)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestWorkOrderRepository_CountByStatus_Coverage 测试按状态计数工单
func TestWorkOrderRepository_CountByStatus_Coverage(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "Success",
			status: "open",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM work_orders WHERE status =`).
					WithArgs("open").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
			},
			wantErr: false,
		},
		{
			name:   "QueryError",
			status: "open",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM work_orders WHERE status =`).
					WillReturnError(errors.New("query error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewWorkOrderRepository(database.NewDBWrapper(db))
			tt.setup(mock)

			count, err := repo.CountByStatus(context.Background(), tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, count, 0)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTelemetryRepository_GetStatsBatch_Coverage 测试批量获取统计数据
func TestTelemetryRepository_GetStatsBatch_Coverage(t *testing.T) {
	// 设备ID测试 - 这是可以测试的边界情况
	t.Run("EmptyDeviceIDs", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewTelemetryRepository(database.NewDBWrapper(db))

		result, err := repo.GetStatsBatch(context.Background(), []string{}, time.Now().Add(-24*time.Hour), time.Now())
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	// 成功测试 - sqlmock不支持PostgreSQL数组参数
	t.Run("Success", func(t *testing.T) {
		t.Skip("sqlmock不支持PostgreSQL ANY($1)数组参数类型")
	})

	// 查询错误测试 - sqlmock不支持PostgreSQL数组参数
	t.Run("QueryError", func(t *testing.T) {
		t.Skip("sqlmock不支持PostgreSQL ANY($1)数组参数类型")
	})
}

// ============================================
// factory.go - 补充测试
// ============================================

// TestRepositoryFactory_GetRBACRepository_Coverage 测试获取RBAC仓库
func TestRepositoryFactory_GetRBACRepository_Coverage(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)

	// 正常情况
	repo, err := factory.GetRBACRepository()
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

// TestRepositoryFactory_GetTenantRepo_Coverage 测试获取租户仓库
func TestRepositoryFactory_GetTenantRepo_Coverage(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)

	repo, err := factory.GetTenantRepo()
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

// TestRepositoryFactory_GetRoleRepo_Coverage 测试获取角色仓库
func TestRepositoryFactory_GetRoleRepo_Coverage(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)

	repo, err := factory.GetRoleRepo()
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

// TestRepositoryFactory_GetPermissionRepo_Coverage 测试获取权限仓库
func TestRepositoryFactory_GetPermissionRepo_Coverage(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	factory := NewRepositoryFactory(db)

	repo, err := factory.GetPermissionRepo()
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

// TestRepositoryFactory_NilDatabase 测试数据库为nil时的错误处理
func TestRepositoryFactory_NilDatabase(t *testing.T) {
	factory := &RepositoryFactory{db: nil}

	repo, err := factory.GetTenantRepo()
	assert.Error(t, err)
	assert.Equal(t, ErrDatabaseNotInitialized, err)
	assert.Nil(t, repo)

	repo2, err := factory.GetRoleRepo()
	assert.Error(t, err)
	assert.Equal(t, ErrDatabaseNotInitialized, err)
	assert.Nil(t, repo2)

	repo3, err := factory.GetPermissionRepo()
	assert.Error(t, err)
	assert.Equal(t, ErrDatabaseNotInitialized, err)
	assert.Nil(t, repo3)

	repo4, err := factory.GetRBACRepository()
	assert.Error(t, err)
	assert.Equal(t, ErrDatabaseNotInitialized, err)
	assert.Nil(t, repo4)
}