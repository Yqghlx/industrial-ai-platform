// Package repository provides common repository patterns
// BE-P2-01: 提取通用 repository 基础方法
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/industrial-ai/platform/pkg/constants"
	"github.com/industrial-ai/platform/pkg/database"
)

// ============================================
// FIX-018: 表名白名单 - 防止 SQL 注入
// ============================================

// allowedTables 是允许在动态 SQL 中使用的表名白名单
// 这些表名已经在代码库中预定义，不接受外部输入
var allowedTables = map[string]bool{
	"users":               true,
	"devices":             true,
	"telemetry_data":      true,
	"alerts":              true,
	"alert_rules":         true,
	"work_orders":         true,
	"notifications":       true,
	"blackbox_records":    true,
	"tenants":             true,
	"roles":               true,
	"permissions":         true,
	"user_roles":          true,
	"role_permissions":    true,
	"agent_task_logs":     true,
	"dashboards":          true,
	"widgets":             true,
	"token_blacklist":     true,
	"user_token_versions": true,
}

// ValidateTableName 验证表名是否在白名单中
// 返回验证后的表名，如果不在白名单中则返回错误
func ValidateTableName(table string) error {
	if table == "" {
		return fmt.Errorf("table name cannot be empty")
	}
	if !allowedTables[table] {
		return fmt.Errorf("invalid table name: %s (not in whitelist)", table)
	}
	return nil
}

// validateTableAndColumn 验证表名和列名
func validateTableAndColumn(table, column string) error {
	if err := ValidateTableName(table); err != nil {
		return err
	}
	if column == "" {
		return fmt.Errorf("column name cannot be empty")
	}
	// 列名只允许字母数字和下划线
	for _, c := range column {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return fmt.Errorf("invalid column name: %s (contains invalid characters)", column)
		}
	}
	return nil
}

// ============================================
// 通用 Repository 接口定义
// ============================================

// BaseRepositoryInterface 定义通用 Repository 接口
type BaseRepositoryInterface interface {
	// Count 返回总数
	Count(ctx context.Context) (int, error)
	// Exists 检查是否存在
	Exists(ctx context.Context, id interface{}) (bool, error)
}

// CRUDRepositoryInterface 定义通用 CRUD 接口
type CRUDRepositoryInterface interface {
	BaseRepositoryInterface
	// Create 创建记录
	Create(ctx context.Context, entity interface{}) error
	// GetByID 根据 ID 获取记录
	GetByID(ctx context.Context, id interface{}) (interface{}, error)
	// Update 更新记录
	Update(ctx context.Context, entity interface{}) error
	// Delete 删除记录
	Delete(ctx context.Context, id interface{}) error
}

// ListRepositoryInterface 定义通用列表接口
type ListRepositoryInterface interface {
	// List 获取列表（分页）
	List(ctx context.Context, page, pageSize int) (interface{}, int, error)
}

// ============================================
// 通用辅助函数
// ============================================

// CalculateOffset 计算分页偏移量
func CalculateOffset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}
	return (page - 1) * pageSize
}

// NormalizePagination 标准化分页参数
func NormalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}
	return page, pageSize
}

// NormalizeLimit 标准化限制参数
func NormalizeLimit(limit int) int {
	if limit < 1 {
		return constants.DefaultLimit
	}
	if limit > constants.MaxLimit {
		return constants.MaxLimit
	}
	return limit
}

// ============================================
// 通用查询辅助方法
// ============================================

// CountQuery 执行通用计数查询
// FIX-018: 添加表名白名单验证防止SQL注入
func CountQuery(ctx context.Context, db database.QueryExecutor, table string) (int, error) {
	if err := ValidateTableName(table); err != nil {
		return 0, err
	}
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	err := db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// ExistsQuery 执行通用存在性查询
// FIX-018: 添加表名和列名白名单验证防止SQL注入
func ExistsQuery(ctx context.Context, db database.QueryExecutor, table, idColumn string, id interface{}) (bool, error) {
	if err := validateTableAndColumn(table, idColumn); err != nil {
		return false, err
	}
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)", table, idColumn)
	err := db.QueryRow(ctx, query, id).Scan(&exists)
	return exists, err
}

// DeleteQuery 执行通用删除查询
// FIX-018: 添加表名和列名白名单验证防止SQL注入
func DeleteQuery(ctx context.Context, db database.QueryExecutor, table, idColumn string, id interface{}) error {
	if err := validateTableAndColumn(table, idColumn); err != nil {
		return err
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, idColumn)
	_, err := db.Exec(ctx, query, id)
	return err
}

// UpdateTimeColumn 更新时间列
func UpdateTimeColumn() time.Time {
	return time.Now()
}

// HandleNoRows 处理 sql.ErrNoRows 错误
func HandleNoRows(err error, entityName string, id interface{}) error {
	if err == sql.ErrNoRows {
		return fmt.Errorf("%s not found: %v", entityName, id)
	}
	return err
}

// ============================================
// 通用事务辅助方法
// ============================================

// WithTransaction 在事务中执行操作
func WithTransaction(ctx context.Context, db database.DatabaseInterface, fn func(tx database.TransactionInterface) error) error {
	if db == nil {
		return fmt.Errorf("database not configured for transactions")
	}

	txHelper := database.NewTransactionHelper(db)
	return txHelper.WithTransaction(ctx, fn)
}

// ============================================
// 通用 Repository 基类
// ============================================

// BaseRepository 提供通用 Repository 实现
type BaseRepository struct {
	db    database.QueryExecutor
	table string
}

// NewBaseRepository 创建通用 Repository
func NewBaseRepository(db database.QueryExecutor, table string) *BaseRepository {
	return &BaseRepository{db: db, table: table}
}

// Count 返回总数
func (r *BaseRepository) Count(ctx context.Context) (int, error) {
	return CountQuery(ctx, r.db, r.table)
}

// Exists 检查是否存在
func (r *BaseRepository) Exists(ctx context.Context, idColumn string, id interface{}) (bool, error) {
	return ExistsQuery(ctx, r.db, r.table, idColumn, id)
}

// Delete 执行删除
func (r *BaseRepository) Delete(ctx context.Context, idColumn string, id interface{}) error {
	return DeleteQuery(ctx, r.db, r.table, idColumn, id)
}
