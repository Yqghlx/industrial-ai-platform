package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/industrial-ai/platform/pkg/errors"
)

// DeviceService handles device business logic
type DeviceService struct {
	deviceRepo *repository.DeviceRepository
	userRepo   *repository.UserRepository
	db         database.DatabaseInterface // for transactions
}

// NewDeviceService creates a new device service
func NewDeviceService(deviceRepo *repository.DeviceRepository, userRepo *repository.UserRepository) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		userRepo:   userRepo,
	}
}

// NewDeviceServiceWithDB creates a device service with database for transactions
func NewDeviceServiceWithDB(deviceRepo *repository.DeviceRepository, userRepo *repository.UserRepository, db database.DatabaseInterface) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		userRepo:   userRepo,
		db:         db,
	}
}

// Create creates a new device
func (s *DeviceService) Create(ctx context.Context, device *model.Device) error {
	// 自动生成 UUID 作为设备 ID（如果前端未提供）
	if device.ID == "" {
		device.ID = uuid.New().String()
	}
	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()
	if device.Status == "" {
		device.Status = "online"
	}
	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return errors.NewDatabaseError(err.Error())
	}
	return nil
}

// GetByID retrieves a device by ID
func (s *DeviceService) GetByID(ctx context.Context, id string) (*model.Device, error) {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewDeviceNotFoundError(id)
	}
	return device, nil
}

// List retrieves devices with pagination
func (s *DeviceService) List(ctx context.Context, page, pageSize int) ([]model.Device, int, error) {
	devices, total, err := s.deviceRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, errors.NewDatabaseError(err.Error())
	}
	return devices, total, nil
}

// Update updates a device
func (s *DeviceService) Update(ctx context.Context, device *model.Device) error {
	// Check if device exists
	_, err := s.deviceRepo.GetByID(ctx, device.ID)
	if err != nil {
		return errors.NewDeviceNotFoundError(device.ID)
	}
	return s.deviceRepo.Update(ctx, device)
}

// Delete removes a device
func (s *DeviceService) Delete(ctx context.Context, id string) error {
	if err := s.deviceRepo.Delete(ctx, id); err != nil {
		return errors.NewDatabaseError(err.Error())
	}
	return nil
}

// UpdateStatus updates device status
func (s *DeviceService) UpdateStatus(ctx context.Context, id, status string) error {
	if err := s.deviceRepo.UpdateStatus(ctx, id, status); err != nil {
		return errors.NewDatabaseError(err.Error())
	}
	return nil
}

// GetDeviceTypeFromID infers device type from ID prefix
func GetDeviceTypeFromID(id string) string {
	if len(id) < 3 {
		return "未知设备"
	}
	prefix := id[:3]
	switch prefix {
	case "CNC":
		return "数控机床"
	case "INJ":
		return "注塑机"
	case "ROB":
		return "工业机器人"
	case "ASM":
		return "装配线"
	case "CNV":
		return "传送带"
	default:
		return "未知设备"
	}
}

// GetDeviceNameFromType generates a device name
func GetDeviceNameFromType(deviceType string) string {
	switch deviceType {
	case "数控机床", "CNC":
		return "数控机床"
	case "注塑机", "INJ":
		return "注塑机"
	case "工业机器人", "ROB":
		return "工业机器人"
	case "装配线", "ASM":
		return "装配线"
	case "传送带", "CNV":
		return "传送带"
	default:
		return "工业设备"
	}
}

// AutoRegisterDevice creates a device if it doesn't exist
func (s *DeviceService) AutoRegisterDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	// Check if device exists
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err == nil {
		return device, nil
	}

	// Auto-register device
	deviceType := GetDeviceTypeFromID(deviceID)
	device = &model.Device{
		ID:          deviceID,
		Name:        GetDeviceNameFromType(deviceType) + " " + deviceID,
		Type:        deviceType,
		Location:    "车间A",
		Status:      "online",
		Description: "自动注册设备",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}

	return device, nil
}

// CreateDeviceWithUser creates a device and associated user in a single transaction
// This demonstrates transaction support across multiple repositories
func (s *DeviceService) CreateDeviceWithUser(ctx context.Context, device *model.Device, user *model.User) error {
	if s.db == nil {
		return errors.NewInternalError("Database not configured for transactions")
	}

	// Use transaction helper
	txHelper := database.NewTransactionHelper(s.db)

	return txHelper.WithTransaction(ctx, func(tx database.TransactionInterface) error {
		// Create device with transaction
		txDeviceRepo := s.deviceRepo.WithTx(tx)
		if err := txDeviceRepo.Create(ctx, device); err != nil {
			return err
		}

		// Create user with transaction
		txUserRepo := s.userRepo.WithTx(tx)
		if err := txUserRepo.Create(ctx, user); err != nil {
			return err
		}

		return nil
	})
}

// GetGraph returns device relationship graph
func (s *DeviceService) GetGraph(ctx context.Context) (map[string]interface{}, error) {
	devices, _, err := s.deviceRepo.List(ctx, 1, 100)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}

	// Build graph structure
	nodes := []map[string]interface{}{}
	links := []map[string]interface{}{}

	for _, d := range devices {
		nodes = append(nodes, map[string]interface{}{
			"id":     d.ID,
			"name":   d.Name,
			"type":   d.Type,
			"status": d.Status,
		})
	}

	// Create sample relationships based on location/type
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if devices[i].Location == devices[j].Location {
				links = append(links, map[string]interface{}{
					"source": devices[i].ID,
					"target": devices[j].ID,
					"type":   "co-located",
				})
			}
		}
	}

	return map[string]interface{}{
		"nodes": nodes,
		"links": links,
	}, nil
}
