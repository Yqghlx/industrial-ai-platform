package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/repository"
)

// ============================================
// 辅助函数
// ============================================

// newTestAlertArchiver 创建测试用 AlertArchiver 实例
func newTestAlertArchiver(mockRepo *repository.MockAlertRepository, config AlertArchiverConfig) *AlertArchiver {
	return NewAlertArchiver(mockRepo, config)
}

// defaultTestConfig 返回测试用默认配置
func defaultTestConfig() AlertArchiverConfig {
	return AlertArchiverConfig{
		ArchiveDaysOld:        30,
		DeleteDaysOld:         90,
		ScheduleIntervalHours: 24,
		EnableArchiving:       true,
	}
}

// ============================================
// NewAlertArchiver 构造函数测试
// ============================================

func TestNewAlertArchiver(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()

	archiver := NewAlertArchiver(mockRepo, cfg)

	// 验证返回值非空
	require.NotNil(t, archiver, "NewAlertArchiver 不应返回 nil")

	// 验证内部字段正确赋值
	assert.Equal(t, mockRepo, archiver.alertRepo, "alertRepo 应与传入参数一致")
	assert.Equal(t, cfg, archiver.config, "config 应与传入参数一致")
	assert.NotNil(t, archiver.stopChan, "stopChan 不应为 nil")
}

// TestNewAlertArchiver_WithCustomConfig 验证自定义配置能正确传递
func TestNewAlertArchiver_WithCustomConfig(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{
		ArchiveDaysOld:        7,
		DeleteDaysOld:         14,
		ScheduleIntervalHours: 1,
		EnableArchiving:       false,
	}

	archiver := NewAlertArchiver(mockRepo, cfg)

	require.NotNil(t, archiver)
	assert.Equal(t, 7, archiver.config.ArchiveDaysOld)
	assert.Equal(t, 14, archiver.config.DeleteDaysOld)
	assert.Equal(t, 1, archiver.config.ScheduleIntervalHours)
	assert.False(t, archiver.config.EnableArchiving)
}

// ============================================
// DefaultAlertArchiverConfig 测试
// ============================================

func TestDefaultAlertArchiverConfig(t *testing.T) {
	cfg := DefaultAlertArchiverConfig()

	assert.Equal(t, 30, cfg.ArchiveDaysOld, "默认归档天数应为 30")
	assert.Equal(t, 90, cfg.DeleteDaysOld, "默认删除天数应为 90")
	assert.Equal(t, 24, cfg.ScheduleIntervalHours, "默认调度间隔应为 24 小时")
	assert.True(t, cfg.EnableArchiving, "默认应启用归档")
}

func TestDefaultAlertArchiverConfig_DeleteAfterArchive(t *testing.T) {
	cfg := DefaultAlertArchiverConfig()

	// 删除天数必须大于归档天数，否则逻辑上不合理
	assert.GreaterOrEqual(t, cfg.DeleteDaysOld, cfg.ArchiveDaysOld,
		"删除天数应大于等于归档天数")
}

// ============================================
// RunManualArchive 测试
// ============================================

// TestRunManualArchive_Success 正常归档和删除都成功
func TestRunManualArchive_Success(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	// 设置 mock 期望：归档 50 条，删除 10 条
	mockRepo.On("ArchiveOldAlerts", ctx, cfg.ArchiveDaysOld).Return(50, nil)
	mockRepo.On("DeleteArchivedAlerts", ctx, cfg.DeleteDaysOld).Return(10, nil)

	archivedCount, deletedCount, err := archiver.RunManualArchive(ctx)

	assert.NoError(t, err, "RunManualArchive 不应返回错误")
	assert.Equal(t, 50, archivedCount, "归档数量应为 50")
	assert.Equal(t, 10, deletedCount, "删除数量应为 10")
	mockRepo.AssertExpectations(t)
}

// TestRunManualArchive_ArchiveError 归档步骤返回错误
func TestRunManualArchive_ArchiveError(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	archiveErr := errors.New("数据库连接失败")
	mockRepo.On("ArchiveOldAlerts", ctx, cfg.ArchiveDaysOld).Return(0, archiveErr)

	archivedCount, deletedCount, err := archiver.RunManualArchive(ctx)

	// 归档失败时应该返回错误，且归档和删除计数均为 0
	assert.ErrorIs(t, err, archiveErr, "应返回归档步骤的错误")
	assert.Equal(t, 0, archivedCount, "归档失败时归档数量应为 0")
	assert.Equal(t, 0, deletedCount, "归档失败时删除数量应为 0")

	// 归档失败后不应调用 DeleteArchivedAlerts
	mockRepo.AssertNotCalled(t, "DeleteArchivedAlerts", ctx, cfg.DeleteDaysOld)
	mockRepo.AssertExpectations(t)
}

// TestRunManualArchive_DeleteError 删除步骤返回错误（归档成功）
func TestRunManualArchive_DeleteError(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	deleteErr := errors.New("删除操作超时")
	mockRepo.On("ArchiveOldAlerts", ctx, cfg.ArchiveDaysOld).Return(30, nil)
	mockRepo.On("DeleteArchivedAlerts", ctx, cfg.DeleteDaysOld).Return(0, deleteErr)

	archivedCount, deletedCount, err := archiver.RunManualArchive(ctx)

	// 删除失败时应返回删除错误，但归档计数仍然保留
	assert.ErrorIs(t, err, deleteErr, "应返回删除步骤的错误")
	assert.Equal(t, 30, archivedCount, "归档成功后的数量应保留")
	assert.Equal(t, 0, deletedCount, "删除失败时删除数量应为 0")
	mockRepo.AssertExpectations(t)
}

// TestRunManualArchive_ZeroResults 归档和删除都返回 0
func TestRunManualArchive_ZeroResults(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	mockRepo.On("ArchiveOldAlerts", ctx, cfg.ArchiveDaysOld).Return(0, nil)
	mockRepo.On("DeleteArchivedAlerts", ctx, cfg.DeleteDaysOld).Return(0, nil)

	archivedCount, deletedCount, err := archiver.RunManualArchive(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 0, archivedCount)
	assert.Equal(t, 0, deletedCount)
	mockRepo.AssertExpectations(t)
}

// TestRunManualArchive_ContextCancelled 上下文取消场景
func TestRunManualArchive_ContextCancelled(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()
	archiver := newTestAlertArchiver(mockRepo, cfg)

	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// mock 仍会被调用，但 repository 实现可能会检查 ctx
	mockRepo.On("ArchiveOldAlerts", ctx, cfg.ArchiveDaysOld).Return(0, context.Canceled)

	archivedCount, deletedCount, err := archiver.RunManualArchive(ctx)

	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 0, archivedCount)
	assert.Equal(t, 0, deletedCount)
	mockRepo.AssertExpectations(t)
}

// ============================================
// GetStatistics 测试
// ============================================

// TestGetStatistics_Success 正常获取统计数据
func TestGetStatistics_Success(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	expectedStats := &repository.AlertStatistics{
		TotalActive:    15,
		TotalResolved:  200,
		TotalArchived:  50,
		TodayTriggered: 3,
		TodayResolved:  5,
		WeekTriggered:  20,
		WeekResolved:   18,
		AvgResolveTime: 3600,
		CriticalCount:  2,
		WarningCount:   10,
	}

	mockRepo.On("GetAlertStatistics", ctx).Return(expectedStats, nil)

	stats, err := archiver.GetStatistics(ctx)

	assert.NoError(t, err, "GetStatistics 不应返回错误")
	require.NotNil(t, stats, "统计数据不应为 nil")
	assert.Equal(t, 15, stats.TotalActive)
	assert.Equal(t, 200, stats.TotalResolved)
	assert.Equal(t, 50, stats.TotalArchived)
	assert.Equal(t, 3, stats.TodayTriggered)
	assert.Equal(t, 5, stats.TodayResolved)
	assert.Equal(t, 2, stats.CriticalCount)
	assert.Equal(t, 10, stats.WarningCount)
	mockRepo.AssertExpectations(t)
}

// TestGetStatistics_Error 仓库层返回错误
func TestGetStatistics_Error(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	dbErr := errors.New("数据库不可用")
	mockRepo.On("GetAlertStatistics", ctx).Return(nil, dbErr)

	stats, err := archiver.GetStatistics(ctx)

	assert.ErrorIs(t, err, dbErr, "应透传仓库层错误")
	assert.Nil(t, stats, "出错时统计数据应为 nil")
	mockRepo.AssertExpectations(t)
}

// ============================================
// Start/Stop 生命周期测试
// ============================================

// TestStartStop_Lifecycle 启动和停止归档器的完整生命周期
func TestStartStop_Lifecycle(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{
		ArchiveDaysOld:        30,
		DeleteDaysOld:         90,
		ScheduleIntervalHours: 24,
		EnableArchiving:       true,
	}
	archiver := newTestAlertArchiver(mockRepo, cfg)

	// runScheduler 启动后会立即执行一次归档周期，需要 mock 相关调用
	// archiveOldAlerts → GetAlertStatistics（runArchiveCycle 的完整流程）
	mockRepo.On("ArchiveOldAlerts", context.Background(), 30).Return(5, nil)
	// DeleteDaysOld (90) > ArchiveDaysOld (30)，所以会调用 DeleteArchivedAlerts
	mockRepo.On("DeleteArchivedAlerts", context.Background(), 90).Return(2, nil)
	mockRepo.On("GetAlertStatistics", context.Background()).Return(&repository.AlertStatistics{}, nil)

	// Start 会启动 goroutine，runScheduler 立即执行一次归档
	archiver.Start()

	// 给 goroutine 足够时间执行初始归档周期
	time.Sleep(100 * time.Millisecond)

	// Stop 关闭 stopChan，使 runScheduler 退出
	archiver.Stop()

	// 等待 goroutine 完全退出
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

// TestStart_Disabled 启用标志为 false 时不启动调度器
func TestStart_Disabled(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{
		ArchiveDaysOld:        30,
		DeleteDaysOld:         90,
		ScheduleIntervalHours: 24,
		EnableArchiving:       false,
	}
	archiver := newTestAlertArchiver(mockRepo, cfg)

	// Start 在 EnableArchiving=false 时不会启动调度器，不会调用任何 repo 方法
	archiver.Start()

	// 给足时间，确认没有 repo 方法被调用
	time.Sleep(50 * time.Millisecond)

	// 不应有任何 mock 调用
	mockRepo.AssertNotCalled(t, "ArchiveOldAlerts")
	mockRepo.AssertNotCalled(t, "DeleteArchivedAlerts")
	mockRepo.AssertNotCalled(t, "GetAlertStatistics")
}

// TestStop_MultipleStopsNotPanic 验证 Stop 只能调用一次（重复 close channel 会 panic）
// 这里只调用一次 Stop，确认正常工作
func TestStop_SingleStop(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()
	archiver := newTestAlertArchiver(mockRepo, cfg)

	// 不调用 Start，直接 Stop —— 此时 stopChan 是打开的
	// close 一个未关闭的 channel 不会 panic
	assert.NotPanics(t, func() {
		archiver.Stop()
	}, "单次 Stop 不应 panic")
}

// ============================================
// runScheduler 内部逻辑的间接测试
// ============================================

// TestRunScheduler_ArchiveError_ContinuesOnError 归档失败后调度器应继续运行
func TestRunScheduler_ArchiveError_ContinuesOnError(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{
		ArchiveDaysOld:        30,
		DeleteDaysOld:         90,
		ScheduleIntervalHours: 24,
		EnableArchiving:       true,
	}
	archiver := newTestAlertArchiver(mockRepo, cfg)

	// 归档失败，但 GetAlertStatistics 仍被调用（runArchiveCycle 继续执行）
	archiveErr := errors.New("归档失败")
	mockRepo.On("ArchiveOldAlerts", context.Background(), 30).Return(0, archiveErr)
	// DeleteDaysOld > ArchiveDaysOld 所以会调用 DeleteArchivedAlerts
	mockRepo.On("DeleteArchivedAlerts", context.Background(), 90).Return(0, nil)
	mockRepo.On("GetAlertStatistics", context.Background()).Return(&repository.AlertStatistics{}, nil)

	// Start 启动调度器，立即执行一次周期
	archiver.Start()
	time.Sleep(100 * time.Millisecond)

	archiver.Stop()
	time.Sleep(50 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

// TestRunScheduler_DeleteSkippedWhenDeleteDaysLessThanArchiveDays
// 当 DeleteDaysOld <= ArchiveDaysOld 时，不应执行删除操作
func TestRunScheduler_DeleteSkippedWhenDeleteDaysLessThanArchiveDays(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{
		ArchiveDaysOld:        90,
		DeleteDaysOld:         30, // 删除天数 < 归档天数
		ScheduleIntervalHours: 24,
		EnableArchiving:       true,
	}
	archiver := newTestAlertArchiver(mockRepo, cfg)

	// 只 mock ArchiveOldAlerts 和 GetAlertStatistics
	mockRepo.On("ArchiveOldAlerts", context.Background(), 90).Return(10, nil)
	mockRepo.On("GetAlertStatistics", context.Background()).Return(&repository.AlertStatistics{}, nil)
	// 注意：不 mock DeleteArchivedAlerts，因为不应被调用

	archiver.Start()
	time.Sleep(100 * time.Millisecond)

	archiver.Stop()
	time.Sleep(50 * time.Millisecond)

	// 验证 DeleteArchivedAlerts 没有被调用
	mockRepo.AssertNotCalled(t, "DeleteArchivedAlerts")
	mockRepo.AssertExpectations(t)
}

// TestRunScheduler_DeleteError_ContinuesOnError 删除失败后仍应继续获取统计
func TestRunScheduler_DeleteError_ContinuesOnError(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{
		ArchiveDaysOld:        30,
		DeleteDaysOld:         90,
		ScheduleIntervalHours: 24,
		EnableArchiving:       true,
	}
	archiver := newTestAlertArchiver(mockRepo, cfg)

	deleteErr := errors.New("删除失败")
	mockRepo.On("ArchiveOldAlerts", context.Background(), 30).Return(5, nil)
	mockRepo.On("DeleteArchivedAlerts", context.Background(), 90).Return(0, deleteErr)
	mockRepo.On("GetAlertStatistics", context.Background()).Return(&repository.AlertStatistics{}, nil)

	archiver.Start()
	time.Sleep(100 * time.Millisecond)

	archiver.Stop()
	time.Sleep(50 * time.Millisecond)

	// 即使删除失败，统计信息仍被获取
	mockRepo.AssertExpectations(t)
}

// TestRunScheduler_StatsError_ContinuesOnError 统计获取失败不影响调度器
func TestRunScheduler_StatsError_ContinuesOnError(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{
		ArchiveDaysOld:        30,
		DeleteDaysOld:         30, // 等于 ArchiveDaysOld，不会执行删除
		ScheduleIntervalHours: 24,
		EnableArchiving:       true,
	}
	archiver := newTestAlertArchiver(mockRepo, cfg)

	statsErr := errors.New("统计获取失败")
	mockRepo.On("ArchiveOldAlerts", context.Background(), 30).Return(5, nil)
	// DeleteDaysOld (30) 不大于 ArchiveDaysOld (30)，所以不会调用 DeleteArchivedAlerts
	mockRepo.On("GetAlertStatistics", context.Background()).Return(nil, statsErr)

	archiver.Start()
	time.Sleep(100 * time.Millisecond)

	archiver.Stop()
	time.Sleep(50 * time.Millisecond)

	// 统计获取失败不应影响调度器运行（不会 panic）
	mockRepo.AssertExpectations(t)
}

// ============================================
// archiveOldAlerts / deleteArchivedAlerts 间接测试
// ============================================

// TestArchiveOldAlerts_Success 验证 archiveOldAlerts 正确传递参数
func TestArchiveOldAlerts_Success(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{ArchiveDaysOld: 60}
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	mockRepo.On("ArchiveOldAlerts", ctx, 60).Return(100, nil)

	count, err := archiver.archiveOldAlerts(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 100, count)
	mockRepo.AssertExpectations(t)
}

// TestArchiveOldAlerts_Error 验证 archiveOldAlerts 正确透传错误
func TestArchiveOldAlerts_Error(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{ArchiveDaysOld: 30}
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	expectedErr := errors.New("连接超时")
	mockRepo.On("ArchiveOldAlerts", ctx, 30).Return(0, expectedErr)

	count, err := archiver.archiveOldAlerts(ctx)

	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, 0, count)
	mockRepo.AssertExpectations(t)
}

// TestDeleteArchivedAlerts_Success 验证 deleteArchivedAlerts 正确传递参数
func TestDeleteArchivedAlerts_Success(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{DeleteDaysOld: 120}
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	mockRepo.On("DeleteArchivedAlerts", ctx, 120).Return(25, nil)

	count, err := archiver.deleteArchivedAlerts(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 25, count)
	mockRepo.AssertExpectations(t)
}

// TestDeleteArchivedAlerts_Error 验证 deleteArchivedAlerts 正确透传错误
func TestDeleteArchivedAlerts_Error(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{DeleteDaysOld: 90}
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	expectedErr := errors.New("权限不足")
	mockRepo.On("DeleteArchivedAlerts", ctx, 90).Return(0, expectedErr)

	count, err := archiver.deleteArchivedAlerts(ctx)

	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, 0, count)
	mockRepo.AssertExpectations(t)
}

// ============================================
// 配置边界场景测试
// ============================================

// TestRunManualArchive_LargeCounts 验证大批量归档/删除的计数正确性
func TestRunManualArchive_LargeCounts(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := defaultTestConfig()
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	mockRepo.On("ArchiveOldAlerts", ctx, 30).Return(10000, nil)
	mockRepo.On("DeleteArchivedAlerts", ctx, 90).Return(5000, nil)

	archivedCount, deletedCount, err := archiver.RunManualArchive(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 10000, archivedCount)
	assert.Equal(t, 5000, deletedCount)
	mockRepo.AssertExpectations(t)
}

// TestRunManualArchive_DifferentDaysConfig 不同天数配置的正确传递
func TestRunManualArchive_DifferentDaysConfig(t *testing.T) {
	mockRepo := new(repository.MockAlertRepository)
	cfg := AlertArchiverConfig{
		ArchiveDaysOld:        7,
		DeleteDaysOld:         14,
		ScheduleIntervalHours: 1,
		EnableArchiving:       true,
	}
	archiver := newTestAlertArchiver(mockRepo, cfg)
	ctx := context.Background()

	mockRepo.On("ArchiveOldAlerts", ctx, 7).Return(20, nil)
	mockRepo.On("DeleteArchivedAlerts", ctx, 14).Return(8, nil)

	archivedCount, deletedCount, err := archiver.RunManualArchive(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 20, archivedCount)
	assert.Equal(t, 8, deletedCount)
	mockRepo.AssertExpectations(t)
}
