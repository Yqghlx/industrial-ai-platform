package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "Admin@123456"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(hashedPassword))
}