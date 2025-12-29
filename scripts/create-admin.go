package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func mainScript() {
	// Change this password as needed
	password := "admin123"

	// Generate bcrypt hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	fmt.Println("=== Admin User Setup ===")
	fmt.Println("\nPassword:", password)
	fmt.Println("\nBcrypt Hash:")
	fmt.Println(string(hash))
	fmt.Println("\n=== MongoDB Command ===")
	fmt.Println("Copy and run this in MongoDB shell:")
	fmt.Println()
	fmt.Printf(`use evote

db.users.insertOne({
  matric_number: "ADMIN001",
  full_name: "Admin User",
  department: "IT",
  faculty: "Administration",
  email: "admin@hallmarkuniversity.edu.ng",
  password_hash: "%s",
  id_card_url: "",
  status: "APPROVED",
  role: "ADMIN",
  has_voted: false,
  created_at: new Date(),
  updated_at: new Date()
})
`, string(hash))

	fmt.Println("\n=== Login Credentials ===")
	fmt.Println("Matric Number: ADMIN001")
	fmt.Println("Password:", password)
	fmt.Println("\nChange the password after first login!")
}
