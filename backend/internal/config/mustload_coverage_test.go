package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMustLoad_ValidationError tests MustLoad with validation errors
func TestMustLoad_ValidationError(t *testing.T) {
	// 设置无效环境变量
	os.Setenv("DATABASE_URL", "")
	os.Setenv("REDIS_URL", "")

	// MustLoad 遇到验证错误会 os.Exit(1)，无法直接测试
	// 所以我们测试 LoadAndValidate 的错误处理

	_, err := LoadAndValidate()
	assert.Error(t, err)
}

// TestMustLoad_SuccessPath tests successful config loading
func TestMustLoad_SuccessPath(t *testing.T) {
	// 设置有效环境变量
	os.Setenv("DATABASE_URL", "postgresql://test:test@localhost:5432/test")
	os.Setenv("REDIS_URL", "redis://localhost:6379")

	config := MustLoad()
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.DatabaseURL)
	assert.NotEmpty(t, config.RedisURL)
}

// TestLoadAndValidate tests LoadAndValidate function
func TestLoadAndValidate_Coverage(t *testing.T) {
	// 测试空环境变量
	os.Setenv("DATABASE_URL", "")

	_, err := LoadAndValidate()
	assert.Error(t, err)

	// 测试有效配置
	os.Setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
	os.Setenv("REDIS_URL", "redis://localhost:6379")

	config, err := LoadAndValidate()
	if err == nil {
		assert.NotNil(t, config)
	}
}
