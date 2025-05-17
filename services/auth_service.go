package services

import (
	"attendance-app/db"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// this GenerateRandomUsername generates a random username for testing
func GenerateRandomUsername() string {
	return fmt.Sprintf("user%d", rand.Intn(10000))
}

// this RegisterUser registers a user and returns secret + provisioning URI, or proceeds to login if user exists
func RegisterUser(username, password string) (string, string, error) {
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=?)", username).Scan(&exists)
	if err != nil {
		return "", "", err
	}
	if exists {
		return "", "", fmt.Errorf("user %s already exists", username) // user exists, no registration needed
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	secret, uri, err := GenerateUserSecret(username)
	if err != nil {
		return "", "", err
	}

	_, err = db.DB.Exec("INSERT INTO users (username, password_hash, secret, role) VALUES (?, ?, ?, ?)", username, string(hashedPassword), secret, "user") // default role is "user"
	if err != nil {
		return "", "", err
	}

	return secret, uri, nil
}

// this VerifyPassword checks password
func VerifyPassword(username, password string) (bool, error) {
	var hashedPassword string
	err := db.DB.QueryRow("SELECT password_hash FROM users WHERE username=?", username).Scan(&hashedPassword)
	if err != nil {
		return false, fmt.Errorf("user not found")
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// this AuthenticateUser verifies password and OTP
func AuthenticateUser(username, password, otp string) (bool, error) {
	passOk, err := VerifyPassword(username, password)
	if err != nil {
		return false, err
	}
	if !passOk {
		return false, fmt.Errorf("invalid password")
	}

	var secret string
	err = db.DB.QueryRow("SELECT secret FROM users WHERE username=?", username).Scan(&secret)
	if err != nil {
		return false, fmt.Errorf("user not found")
	}

	if !VerifyTOTP(secret, otp) {
		return false, fmt.Errorf("invalid OTP")
	}
	return true, nil
}

// this CheckRole checks the role of the user (admin or user)
func CheckRole(username string) (string, error) {
	var role string
	err := db.DB.QueryRow("SELECT role FROM users WHERE username=?", username).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve user role")
	}
	return role, nil
}

// this ResetPassword allows admin to reset user password
func ResetPassword(adminUsername, targetUsername, newPassword string) error {
	// Check if the admin has the correct role
	role, err := CheckRole(adminUsername)
	if err != nil || role != "admin" {
		return fmt.Errorf("unauthorized: admin privileges required")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update the user's password
	_, err = db.DB.Exec("UPDATE users SET password_hash=? WHERE username=?", hashedPassword, targetUsername)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

// this ChangeRole allows admin to change the role of a user
func ChangeRole(adminUsername, targetUsername, newRole string) error {
	// Check if the admin has the correct role
	role, err := CheckRole(adminUsername)
	if err != nil || role != "admin" {
		return fmt.Errorf("unauthorized: admin privileges required")
	}

	// Update the user's role
	_, err = db.DB.Exec("UPDATE users SET role=? WHERE username=?", newRole, targetUsername)
	if err != nil {
		return fmt.Errorf("failed to change role: %w", err)
	}

	return nil
}

// this DeleteUser allows admin to delete a user
func DeleteUser(adminUsername, targetUsername string) error {
	// Check if the admin has the correct role
	role, err := CheckRole(adminUsername)
	if err != nil || role != "admin" {
		return fmt.Errorf("unauthorized: admin privileges required")
	}

	// Delete the user
	_, err = db.DB.Exec("DELETE FROM users WHERE username=?", targetUsername)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// this RecordAttendance logs attendance for user
func RecordAttendance(username string) error {
	_, err := db.DB.Exec("INSERT INTO attendance (username, timestamp, status) VALUES (?, ?, ?)", username, time.Now(), "present")
	return err
}

// this LogAudit logs events like login attempts and attendance actions
func LogAudit(username, eventType string, success bool, ip string) error {
	_, err := db.DB.Exec("INSERT INTO audit_logs (username, event_type, success, event_time, ip_address) VALUES (?, ?, ?, ?, ?)", username, eventType, success, time.Now(), ip)
	return err
}

// this ListUsers prints all registered usernames (admins only)
func ListUsers() error {
	rows, err := db.DB.Query("SELECT username FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()
	fmt.Println("\n📝 Registered Users:")
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return err
		}
		fmt.Println(" -", username)
	}
	fmt.Println()
	return nil
}
