package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/industrial-ai/platform/internal/model"
)

// DeviceRepositoryInterface defines the interface for device repository
type DeviceRepositoryInterface interface {
	Create(ctx context.Context, device *model.Device) error
	GetByID(ctx context.Context, id string) (*model.Device, error)
	List(ctx context.Context, page, pageSize int) ([]model.Device, int, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
	Count(ctx context.Context) (int, error)
}

// DeviceRepository handles device data access
type DeviceRepository struct {
	db *sql.DB
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(db *sql.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
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
	_, err := r.db.ExecContext(ctx, query,
		device.ID, device.Name, device.Type, device.Location,
		device.Status, device.Description, device.CreatedAt, device.UpdatedAt,
	)
	return err
}

// GetByID retrieves a device by ID
func (r *DeviceRepository) GetByID(ctx context.Context, id string) (*model.Device, error) {
	query := `
		SELECT id, name, type, location, status, description, created_at, updated_at
		FROM devices WHERE id = $1
	`
	device := &model.Device{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	query := `
		SELECT id, name, type, location, status, description, created_at, updated_at
		FROM devices ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
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
	_, err := r.db.ExecContext(ctx, query,
		device.Name, device.Type, device.Location, device.Status,
		device.Description, device.UpdatedAt, device.ID,
	)
	return err
}

// Delete removes a device
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM devices WHERE id = $1", id)
	return err
}

// UpdateStatus updates only the status of a device
func (r *DeviceRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE devices SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now(), id,
	)
	return err
}

// Count returns total device count
func (r *DeviceRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices").Scan(&count)
	return count, err
}

// UserRepository handles user data access
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, password_hash, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query,
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
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
	err := r.db.QueryRowContext(ctx, query, username).Scan(
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
	err := r.db.QueryRowContext(ctx, query, email).Scan(
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
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, username, password_hash, email, role, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Password, &u.Email,
			&u.Role, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
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
	_, err := r.db.ExecContext(ctx, query,
		user.Username, user.Email, user.Role, user.UpdatedAt, user.ID,
	)
	return err
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, id int, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3",
		passwordHash, time.Now(), id,
	)
	return err
}

// Delete removes a user
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

// Count returns total user count
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// GetTokenVersion 获取用户的 Token 版本号
func (r *UserRepository) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	var version int
	err := r.db.QueryRowContext(ctx,
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
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET token_version = token_version + 1, updated_at = $1 WHERE id = $2",
		time.Now(), userID,
	)
	return err
}
