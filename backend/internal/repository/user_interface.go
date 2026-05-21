package repository

import (
	"context"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
)

// UserRepositoryInterface defines the interface for user repository
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, page, pageSize int) ([]model.User, int, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int) error
	UpdatePassword(ctx context.Context, id int, passwordHash string) error
	GetTokenVersion(ctx context.Context, userID int) (int, error)
	UpdateTokenVersion(ctx context.Context, userID int) error
	WithTx(tx database.TransactionInterface) UserRepositoryInterface
}
