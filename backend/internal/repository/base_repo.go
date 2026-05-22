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
func CountQuery(ctx context.Context, db database.QueryExecutor, table string) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	err := db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// ExistsQuery 执行通用存在性查询
func ExistsQuery(ctx context.Context, db database.QueryExecutor, table, idColumn string, id interface{}) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)", table, idColumn)
	err := db.QueryRow(ctx, query, id).Scan(&exists)
	return exists, err
}

// DeleteQuery 执行通用删除查询
func DeleteQuery(ctx context.Context, db database.QueryExecutor, table, idColumn string, id interface{}) error {
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