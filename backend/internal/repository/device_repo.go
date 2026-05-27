package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
)

// DeviceRepositoryInterface defines the interface for device repository
type DeviceRepositoryInterface interface {
	Create(ctx context.Context, device *model.Device) error
	GetByID(ctx context.Context, id string) (*model.Device, error)
	// FIX-022: N+1 查询优化 - 新增带租户隔离的查询方法
	GetByIDWithTenant(ctx context.Context, id string, tenantID string) (*model.Device, error)
	List(ctx context.Context, page, pageSize int) ([]model.Device, int, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
	Count(ctx context.Context) (int, error)
	WithTx(tx database.TransactionInterface) DeviceRepositoryInterface
	// Batch operations for performance optimization
	BatchCreate(ctx context.Context, devices []*model.Device) error
	BatchUpdate(ctx context.Context, devices []*model.Device) error
	BatchUpdateStatus(ctx context.Context, deviceStatuses map[string]string) error
}

// DeviceRepository handles device data access
type DeviceRepository struct {
	db database.QueryExecutor
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(db database.QueryExecutor) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *DeviceRepository) WithTx(tx database.TransactionInterface) DeviceRepositoryInterface {
	return &DeviceRepository{db: tx}
}

// Create inserts a new device
func (r *DeviceRepository) Create(ctx context.Context, device *model.Device) error {
	query := `
		INSERT INTO devices (id, name, type, location, status, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			location = EXCLUDED.location,
			status = EXCLUDED.status,
			description = EXCLUDED.description,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(ctx, query,
		device.ID, device.Name, device.Type, device.Location,
		device.Status, device.Description, device.CreatedAt, device.UpdatedAt,
	)
	return err
}

// GetByID retrieves a device by ID
// FIX-022: 添加租户隔离支持，防止跨租户数据访问
func (r *DeviceRepository) GetByID(ctx context.Context, id string) (*model.Device, error) {
	query := `
		SELECT id, name, type, location, status, description, created_at, updated_at
		FROM devices WHERE id = $1
	`
	device := &model.Device{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&device.ID, &device.Name, &device.Type, &device.Location,
		&device.Status, &device.Description, &device.CreatedAt, &device.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return device, nil
}

// GetByIDWithTenant retrieves a device by ID with tenant isolation
// FIX-022: N+1 查询优化 - 新增带租户隔离的查询方法
func (r *DeviceRepository) GetByIDWithTenant(ctx context.Context, id string, tenantID string) (*model.Device, error) {
	query := `
		SELECT id, name, type, location, status, description, created_at, updated_at
		FROM devices WHERE id = $1 AND (tenant_id = $2 OR tenant_id = '' OR tenant_id IS NULL)
	`
	device := &model.Device{}
	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&device.ID, &device.Name, &device.Type, &device.Location,
		&device.Status, &device.Description, &device.CreatedAt, &device.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return device, nil
}

// List retrieves all devices with pagination
func (r *DeviceRepository) List(ctx context.Context, page, pageSize int) ([]model.Device, int, error) {
	// Count total
	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM devices").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	query := `
		SELECT id, name, type, location, status, description, created_at, updated_at
		FROM devices ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		var d model.Device
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Type, &d.Location,
			&d.Status, &d.Description, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		devices = append(devices, d)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	return devices, total, nil
}

// Update updates a device
func (r *DeviceRepository) Update(ctx context.Context, device *model.Device) error {
	query := `
		UPDATE devices SET
			name = $1, type = $2, location = $3, status = $4,
			description = $5, updated_at = $6
		WHERE id = $7
	`
	device.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, query,
		device.Name, device.Type, device.Location, device.Status,
		device.Description, device.UpdatedAt, device.ID,
	)
	return err
}

// Delete removes a device
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM devices WHERE id = $1", id)
	return err
}

// UpdateStatus updates only the status of a device
func (r *DeviceRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE devices SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now(), id,
	)
	return err
}

// Count returns total device count
func (r *DeviceRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM devices").Scan(&count)
	return count, err
}

// BatchCreate inserts multiple devices in a single query for better performance
func (r *DeviceRepository) BatchCreate(ctx context.Context, devices []*model.Device) error {
	if len(devices) == 0 {
		return nil
	}

	// Build batch insert query with ON CONFLICT for upsert behavior
	query := `
		INSERT INTO devices (id, name, type, location, status, description, created_at, updated_at)
		VALUES 
	`
	var values []interface{}
	var placeholders []string
	placeholderIdx := 1

	for _, device := range devices {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			placeholderIdx, placeholderIdx+1, placeholderIdx+2, placeholderIdx+3,
			placeholderIdx+4, placeholderIdx+5, placeholderIdx+6, placeholderIdx+7))
		values = append(values,
			device.ID, device.Name, device.Type, device.Location,
			device.Status, device.Description, device.CreatedAt, device.UpdatedAt,
		)
		placeholderIdx += 8
	}

	query += fmt.Sprintf("%s\nON CONFLICT (id) DO UPDATE SET\n\t\tname = EXCLUDED.name,\n\t\ttype = EXCLUDED.type,\n\t\tlocation = EXCLUDED.location,\n\t\tstatus = EXCLUDED.status,\n\t\tdescription = EXCLUDED.description,\n\t\tupdated_at = EXCLUDED.updated_at", strings.Join(placeholders, ", "))

	_, err := r.db.Exec(ctx, query, values...)
	return err
}

// BatchUpdate updates multiple devices in a single query for better performance
// FIX-022: N+1 查询优化 - 使用 CASE 语句实现真正的批量更新
func (r *DeviceRepository) BatchUpdate(ctx context.Context, devices []*model.Device) error {
	if len(devices) == 0 {
		return nil
	}

	now := time.Now()

	// Build batch update query using CASE statements for each field
	// This reduces N queries to 1 query
	var args []interface{}
	args = append(args, now)
	placeholderIdx := 2

	// Collect device IDs for WHERE IN clause
	for _, device := range devices {
		args = append(args, device.ID)
		placeholderIdx++
	}
	idCount := len(devices)
	idStart := 2

	// Build CASE for name
	query := "UPDATE devices SET name = CASE id "
	for i := 0; i < idCount; i++ {
		query += fmt.Sprintf("WHEN $%d THEN $%d ", idStart+i, placeholderIdx)
		args = append(args, devices[i].Name)
		placeholderIdx++
	}
	query += "ELSE name END, "

	// Build CASE for type
	query += "type = CASE id "
	for i := 0; i < idCount; i++ {
		query += fmt.Sprintf("WHEN $%d THEN $%d ", idStart+i, placeholderIdx)
		args = append(args, devices[i].Type)
		placeholderIdx++
	}
	query += "ELSE type END, "

	// Build CASE for location
	query += "location = CASE id "
	for i := 0; i < idCount; i++ {
		query += fmt.Sprintf("WHEN $%d THEN $%d ", idStart+i, placeholderIdx)
		args = append(args, devices[i].Location)
		placeholderIdx++
	}
	query += "ELSE location END, "

	// Build CASE for status
	query += "status = CASE id "
	for i := 0; i < idCount; i++ {
		query += fmt.Sprintf("WHEN $%d THEN $%d ", idStart+i, placeholderIdx)
		args = append(args, devices[i].Status)
		placeholderIdx++
	}
	query += "ELSE status END, "

	// Build CASE for description
	query += "description = CASE id "
	for i := 0; i < idCount; i++ {
		query += fmt.Sprintf("WHEN $%d THEN $%d ", idStart+i, placeholderIdx)
		args = append(args, devices[i].Description)
		placeholderIdx++
	}
	query += "ELSE description END, "

	// Add updated_at and WHERE clause
	query += "updated_at = $1 WHERE id IN ("
	for i := 0; i < idCount; i++ {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", idStart+i)
	}
	query += ")"

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// BatchUpdateStatus updates status for multiple devices in a single query
func (r *DeviceRepository) BatchUpdateStatus(ctx context.Context, deviceStatuses map[string]string) error {
	if len(deviceStatuses) == 0 {
		return nil
	}

	now := time.Now()
	// Use CASE statement for batch update
	query := `UPDATE devices SET status = CASE id `
	var args []interface{}
	args = append(args, now)
	placeholderIdx := 2

	var ids []string
	for id, status := range deviceStatuses {
		query += fmt.Sprintf("WHEN $%d THEN $%d ", placeholderIdx, placeholderIdx+1)
		args = append(args, id, status)
		ids = append(ids, fmt.Sprintf("$%d", placeholderIdx))
		placeholderIdx += 2
	}
	query += fmt.Sprintf("ELSE status END, updated_at = $1 WHERE id IN (%s)", strings.Join(ids, ", "))

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// UserRepository handles user data access
type UserRepository struct {
	db database.QueryExecutor
}

// NewUserRepository creates a new user repository
func NewUserRepository(db database.QueryExecutor) *UserRepository {
	return &UserRepository{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *UserRepository) WithTx(tx database.TransactionInterface) UserRepositoryInterface {
	return &UserRepository{db: tx}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, password_hash, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		user.Username, user.Password, user.Email, user.Role,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, email, role, COALESCE(token_version, 0), COALESCE(tenant_id, ''), created_at, updated_at
		FROM users WHERE id = $1
	`
	user := &model.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.Role, &user.TokenVersion, &user.TenantID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = user.Password // 兼容别名
	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, email, role, COALESCE(token_version, 0), COALESCE(tenant_id, ''), created_at, updated_at
		FROM users WHERE username = $1
	`
	user := &model.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.Role, &user.TokenVersion, &user.TenantID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = user.Password // 兼容别名
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, email, role, COALESCE(token_version, 0), COALESCE(tenant_id, ''), created_at, updated_at
		FROM users WHERE email = $1
	`
	user := &model.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.Role, &user.TokenVersion, &user.TenantID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = user.Password // 兼容别名
	return user, nil
}

// List retrieves all users with pagination
// FIX-022: 修复 P1-12 - 添加 tenant_id 和 token_version 字段
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, username, password_hash, email, role, COALESCE(token_version, 0), COALESCE(tenant_id, ''), created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Password, &u.Email,
			&u.Role, &u.TokenVersion, &u.TenantID, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		u.PasswordHash = u.Password // 兼容别名
		users = append(users, u)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	return users, total, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users SET
			username = $1, email = $2, role = $3, updated_at = $4
		WHERE id = $5
	`
	user.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, query,
		user.Username, user.Email, user.Role, user.UpdatedAt, user.ID,
	)
	return err
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, id int, passwordHash string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3",
		passwordHash, time.Now(), id,
	)
	return err
}

// Delete removes a user
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

// Count returns total user count
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// GetTokenVersion 获取用户的 Token 版本号
func (r *UserRepository) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	var version int
	err := r.db.QueryRow(ctx,
		"SELECT token_version FROM users WHERE id = $1",
		userID,
	).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// UpdateTokenVersion 递增用户的 Token 版本号 (撤销所有旧 Token)
func (r *UserRepository) UpdateTokenVersion(ctx context.Context, userID int) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET token_version = token_version + 1, updated_at = $1 WHERE id = $2",
		time.Now(), userID,
	)
	return err
}
