package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

func main() {
	// 从环境变量读取数据库连接字符串
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL 环境变量必须设置，例如：postgres://user:pass@localhost:5432/industrial_ai?sslmode=disable")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 从命令行参数读取新密码，默认 Admin@123456
	password := "Admin@123456"
	if len(os.Args) > 1 {
		password = os.Args[1]
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 更新管理员密码
	query := `UPDATE users SET password_hash = $1 WHERE username = 'admin'`
	result, err := db.ExecContext(context.Background(), query, string(hashedPassword))
	if err != nil {
		log.Fatalf("Failed to update password: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Fatal("未找到 admin 用户，请确认数据库已初始化")
	}

	fmt.Printf("Admin 密码重置成功。影响行数: %d\n", rowsAffected)
}
