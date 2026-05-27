package service

import (
	"context"
	"time"

	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// AlertArchiverConfig 告警归档配置
type AlertArchiverConfig struct {
	// ArchiveDaysOld 归档超过此天数的已解决告警
	ArchiveDaysOld int
	// DeleteDaysOld 删除超过此天数的归档告警
	DeleteDaysOld int
	// ScheduleInterval 调度间隔（小时）
	ScheduleIntervalHours int
	// EnableArchiving 是否启用归档
	EnableArchiving bool
}

// DefaultAlertArchiverConfig 默认配置
func DefaultAlertArchiverConfig() AlertArchiverConfig {
	return AlertArchiverConfig{
		ArchiveDaysOld:        30,  // 30天后归档
		DeleteDaysOld:         90,  // 90天后删除归档
		ScheduleIntervalHours: 24,  // 每天执行一次
		EnableArchiving:       true,
	}
}

// AlertArchiver 告警历史归档器
type AlertArchiver struct {
	alertRepo repository.AlertRepositoryInterface
	config    AlertArchiverConfig
	stopChan  chan struct{}
}

// NewAlertArchiver 创建告警归档器
func NewAlertArchiver(alertRepo AlertRepositoryInterface, config AlertArchiverConfig) *AlertArchiver {
	return &AlertArchiver{
		alertRepo: alertRepo,
		config:    config,
		stopChan:  make(chan struct{}),
	}
}

// Start 启动归档调度
func (a *AlertArchiver) Start() {
	if !a.config.EnableArchiving {
		logger.L().Info("Alert archiving is disabled")
		return
	}

	// 启动调度器
	go a.runScheduler()
	
	logger.L().Info("Alert archiver started",
		zap.Int("archive_days", a.config.ArchiveDaysOld),
		zap.Int("delete_days", a.config.DeleteDaysOld),
		zap.Int("interval_hours", a.config.ScheduleIntervalHours),
	)
}

// Stop 停止归档调度
func (a *AlertArchiver) Stop() {
	close(a.stopChan)
	logger.L().Info("Alert archiver stopped")
}

// runScheduler 运行调度器
func (a *AlertArchiver) runScheduler() {
	// 立即执行一次归档
	a.runArchiveCycle()
	
	// 定时执行
	ticker := time.NewTicker(time.Duration(a.config.ScheduleIntervalHours) * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			a.runArchiveCycle()
		case <-a.stopChan:
			return
		}
	}
}

// runArchiveCycle 执行归档周期
func (a *AlertArchiver) runArchiveCycle() {
	ctx := context.Background()
	startTime := time.Now()
	
	// 1. 归档旧告警
	archivedCount, err := a.archiveOldAlerts(ctx)
	if err != nil {
		logger.L().Error("Failed to archive old alerts",
			zap.Error(err),
		)
	} else {
		logger.L().Info("Archived old alerts",
			zap.Int("count", archivedCount),
			zap.Int("days_old", a.config.ArchiveDaysOld),
		)
	}
	
	// 2. 删除过期归档
	if a.config.DeleteDaysOld > a.config.ArchiveDaysOld {
		deletedCount, err := a.deleteArchivedAlerts(ctx)
		if err != nil {
			logger.L().Error("Failed to delete archived alerts",
				zap.Error(err),
			)
		} else {
			logger.L().Info("Deleted archived alerts",
				zap.Int("count", deletedCount),
				zap.Int("days_old", a.config.DeleteDaysOld),
			)
		}
	}
	
	// 3. 记录统计信息
	stats, err := a.alertRepo.GetAlertStatistics(ctx)
	if err != nil {
		logger.L().Warn("Failed to get alert statistics",
			zap.Error(err),
		)
	} else {
		logger.L().Info("Alert statistics",
			zap.Int("active", stats.TotalActive),
			zap.Int("resolved", stats.TotalResolved),
			zap.Int("archived", stats.TotalArchived),
			zap.Int("today_triggered", stats.TodayTriggered),
			zap.Int("today_resolved", stats.TodayResolved),
		)
	}
	
	duration := time.Since(startTime)
	logger.L().Info("Archive cycle completed",
		zap.Duration("duration", duration),
	)
}

// archiveOldAlerts 归档旧告警
func (a *AlertArchiver) archiveOldAlerts(ctx context.Context) (int, error) {
	return a.alertRepo.ArchiveOldAlerts(ctx, a.config.ArchiveDaysOld)
}

// deleteArchivedAlerts 删除归档告警
func (a *AlertArchiver) deleteArchivedAlerts(ctx context.Context) (int, error) {
	return a.alertRepo.DeleteArchivedAlerts(ctx, a.config.DeleteDaysOld)
}

// RunManualArchive 手动执行归档（用于测试或紧急情况）
func (a *AlertArchiver) RunManualArchive(ctx context.Context) (int, int, error) {
	archivedCount, err := a.alertRepo.ArchiveOldAlerts(ctx, a.config.ArchiveDaysOld)
	if err != nil {
		return 0, 0, err
	}
	
	deletedCount, err := a.alertRepo.DeleteArchivedAlerts(ctx, a.config.DeleteDaysOld)
	if err != nil {
		return archivedCount, 0, err
	}
	
	return archivedCount, deletedCount, nil
}

// GetStatistics 获取告警统计
func (a *AlertArchiver) GetStatistics(ctx context.Context) (*repository.AlertStatistics, error) {
	return a.alertRepo.GetAlertStatistics(ctx)
}