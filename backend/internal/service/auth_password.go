package service

import "golang.org/x/crypto/bcrypt"

// SEC-HIGH-04: bcrypt 成本因子
// 使用成本因子 12 以提供更强的密码哈希保护 (2026年安全标准)
// 注意: 成本因子每增加1，计算时间翻倍。12是一个安全与性能的平衡点。
const BcryptCost = 12

// HashPassword hashes a password using bcrypt with secure cost factor
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	return string(bytes), err
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
