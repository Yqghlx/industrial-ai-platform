package service

import (
	"context"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

// ============================================
// Phase 3补充: 创建缺失的Service层
// ============================================

// WorkOrderService 工单服务
type WorkOrderService struct {
	workOrderRepo repository.WorkOrderRepositoryInterface
}

// NewWorkOrderService 创建工单服务
func NewWorkOrderService(repo repository.WorkOrderRepositoryInterface) *WorkOrderService {
	return &WorkOrderService{workOrderRepo: repo}
}

// FIX-019: 添加 Context 超时设置
func (s *WorkOrderService) List(ctx context.Context, status, deviceID string, page, pageSize int) ([]model.WorkOrder, int, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.workOrderRepo.List(ctx, status, deviceID, page, pageSize)
}

// FIX-019: 添加 Context 超时设置
func (s *WorkOrderService) Create(ctx context.Context, order *model.WorkOrder) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.workOrderRepo.Create(ctx, order)
}

// FIX-019: 添加 Context 超时设置
func (s *WorkOrderService) UpdateStatus(ctx context.Context, id int, status string) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.workOrderRepo.UpdateStatus(ctx, id, status)
}

// FIX-019: 添加 Context 超时设置
func (s *WorkOrderService) GetByID(ctx context.Context, id int) (*model.WorkOrder, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.workOrderRepo.GetByID(ctx, id)
}

// NotificationService 通知服务
type NotificationService struct {
	notificationRepo repository.NotificationRepositoryInterface
}

// NewNotificationService 创建通知服务
func NewNotificationService(repo repository.NotificationRepositoryInterface) *NotificationService {
	return &NotificationService{notificationRepo: repo}
}

// FIX-019: 添加 Context 超时设置
func (s *NotificationService) List(ctx context.Context, status string, page, pageSize int) ([]model.Notification, int, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	unreadOnly := status == "unread"
	return s.notificationRepo.List(ctx, "", unreadOnly, page, pageSize)
}

// FIX-019: 添加 Context 超时设置
func (s *NotificationService) MarkRead(ctx context.Context, id int) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.notificationRepo.MarkRead(ctx, id)
}

// FIX-019: 添加 Context 超时设置
func (s *NotificationService) Create(ctx context.Context, n *model.Notification) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.notificationRepo.Create(ctx, n)
}

// BlackBoxService 黑匣子服务
type BlackBoxService struct {
	blackBoxRepo repository.BlackBoxRepositoryInterface
}

// NewBlackBoxService 创建黑匣子服务
func NewBlackBoxService(repo repository.BlackBoxRepositoryInterface) *BlackBoxService {
	return &BlackBoxService{blackBoxRepo: repo}
}

// FIX-019: 添加 Context 超时设置
func (s *BlackBoxService) List(ctx context.Context, deviceID string, page, pageSize int) ([]model.BlackBoxRecord, int, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.blackBoxRepo.List(ctx, deviceID, page, pageSize)
}

// FIX-019: 添加 Context 超时设置
func (s *BlackBoxService) Create(ctx context.Context, record *model.BlackBoxRecord) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.blackBoxRepo.Create(ctx, record)
}

// TelemetryService补充方法
// FIX-019: 添加 Context 超时设置
func (s *TelemetryService) GetLatestByDevice(ctx context.Context, deviceID string, limit int) ([]model.TelemetryData, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.telemetryRepo.GetByDeviceID(ctx, deviceID, time.Now().AddDate(0, 0, -7), time.Now(), limit)
}

// 接口实现验证 (编译时检查)
var _ WorkOrderServiceInterface = (*WorkOrderService)(nil)
var _ NotificationServiceInterface = (*NotificationService)(nil)
var _ BlackBoxServiceInterface = (*BlackBoxService)(nil)
