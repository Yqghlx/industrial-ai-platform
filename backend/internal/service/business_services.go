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

func (s *WorkOrderService) List(ctx context.Context, status, deviceID string, page, pageSize int) ([]model.WorkOrder, int, error) {
	return s.workOrderRepo.List(ctx, status, deviceID, page, pageSize)
}

func (s *WorkOrderService) Create(ctx context.Context, order *model.WorkOrder) error {
	return s.workOrderRepo.Create(ctx, order)
}

func (s *WorkOrderService) UpdateStatus(ctx context.Context, id int, status string) error {
	return s.workOrderRepo.UpdateStatus(ctx, id, status)
}

func (s *WorkOrderService) GetByID(ctx context.Context, id int) (*model.WorkOrder, error) {
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

func (s *NotificationService) List(ctx context.Context, status string, page, pageSize int) ([]model.Notification, int, error) {
	unreadOnly := status == "unread"
	return s.notificationRepo.List(ctx, "", unreadOnly, page, pageSize)
}

func (s *NotificationService) MarkRead(ctx context.Context, id int) error {
	return s.notificationRepo.MarkRead(ctx, id)
}

func (s *NotificationService) Create(ctx context.Context, n *model.Notification) error {
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

func (s *BlackBoxService) List(ctx context.Context, deviceID string, page, pageSize int) ([]model.BlackBoxRecord, int, error) {
	return s.blackBoxRepo.List(ctx, deviceID, page, pageSize)
}

func (s *BlackBoxService) Create(ctx context.Context, record *model.BlackBoxRecord) error {
	return s.blackBoxRepo.Create(ctx, record)
}

// TelemetryService补充方法
func (s *TelemetryService) GetLatestByDevice(ctx context.Context, deviceID string, limit int) ([]model.TelemetryData, error) {
	return s.telemetryRepo.GetByDeviceID(ctx, deviceID, time.Now().AddDate(0, 0, -7), time.Now(), limit)
}

// 接口实现验证 (编译时检查)
var _ WorkOrderServiceInterface = (*WorkOrderService)(nil)
var _ NotificationServiceInterface = (*NotificationService)(nil)
var _ BlackBoxServiceInterface = (*BlackBoxService)(nil)
