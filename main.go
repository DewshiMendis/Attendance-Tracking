package main

import (
	"attendance-app/db"
	"attendance-app/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	db.InitDB()      // Initialize the SQLite database connection
	defer db.Close() // Close the DB connection when the program ends

	// Set up router
	r := mux.NewRouter()

	// Serve static files (HTML, CSS, JS)
	staticDir := http.Dir("./static")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(staticDir)))

	// Serve QR code images
	r.PathPrefix("/qrcodes/").Handler(http.StripPrefix("/qrcodes/", http.FileServer(http.Dir("."))))

	// API endpoints
	r.HandleFunc("/api/register", handleRegister).Methods("POST")
	r.HandleFunc("/api/login", handleLogin).Methods("POST")
	r.HandleFunc("/api/attendance", handleRecordAttendance).Methods("POST")
	r.HandleFunc("/api/attendance/dates", handleGetAttendanceDates).Methods("GET")
	r.HandleFunc("/api/admin/reset-password", handleResetPassword).Methods("POST")
	r.HandleFunc("/api/admin/change-role", handleChangeRole).Methods("POST")
	r.HandleFunc("/api/admin/delete-user", handleDeleteUser).Methods("POST")
	r.HandleFunc("/api/admin/list-users", handleListUsers).Methods("GET")
	r.HandleFunc("/api/verify-otp", handleVerifyOTP).Methods("POST")

	// Serve index.html for the root path
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// Start server
	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	OTP      string `json:"otp"`
}

type VerifyOTPRequest struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
	Secret   string `json:"secret"`
}

type ResetPasswordRequest struct {
	AdminUsername  string `json:"adminUsername"`
	TargetUsername string `json:"targetUsername"`
	NewPassword    string `json:"newPassword"`
}

type ChangeRoleRequest struct {
	AdminUsername  string `json:"adminUsername"`
	TargetUsername string `json:"targetUsername"`
	NewRole        string `json:"newRole"`
}

type DeleteUserRequest struct {
	AdminUsername  string `json:"adminUsername"`
	TargetUsername string `json:"targetUsername"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	secret, uri, err := services.RegisterUser(req.Username, req.Password)
	if err != nil {
		if err.Error() == fmt.Sprintf("user %s already exists", req.Username) {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate QR code PNG
	qrFilename := fmt.Sprintf("%s_qrcode.png", req.Username)
	if err := qrcode.WriteFile(uri, qrcode.Medium, 256, qrFilename); err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	// Respond with success and QR code URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Registered successfully",
		"secret":    secret,
		"qrCodeUrl": fmt.Sprintf("/qrcodes/%s", qrFilename),
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	success, err := services.AuthenticateUser(req.Username, req.Password, req.OTP)
	if err != nil || !success {
		services.LogAudit(req.Username, "login_attempt", false, r.RemoteAddr)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	services.LogAudit(req.Username, "login_attempt", true, r.RemoteAddr)
	role, err := services.CheckRole(req.Username)
	if err != nil {
		http.Error(w, "Error checking role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Authentication successful",
		"role":    role,
	})
}

func handleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if services.VerifyTOTP(req.Secret, req.OTP) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "OTP verified successfully",
		})
		return
	}

	http.Error(w, "Invalid OTP", http.StatusUnauthorized)
}

func handleRecordAttendance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := services.RecordAttendance(req.Username); err != nil {
		services.LogAudit(req.Username, "attendance_record", false, r.RemoteAddr)
		http.Error(w, "Failed to record attendance", http.StatusInternalServerError)
		return
	}

	services.LogAudit(req.Username, "attendance_record", true, r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Attendance recorded successfully",
	})
}

func handleGetAttendanceDates(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	rows, err := db.DB.Query("SELECT timestamp FROM attendance WHERE username = ? AND status = 'present'", username)
	if err != nil {
		http.Error(w, "Failed to fetch attendance dates", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var timestamp time.Time
		if err := rows.Scan(&timestamp); err != nil {
			http.Error(w, "Error scanning dates", http.StatusInternalServerError)
			return
		}
		dates = append(dates, timestamp.Format("2006-01-02"))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"dates": dates,
	})
}

func handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := services.ResetPassword(req.AdminUsername, req.TargetUsername, req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successfully",
	})
}

func handleChangeRole(w http.ResponseWriter, r *http.Request) {
	var req ChangeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := services.ChangeRole(req.AdminUsername, req.TargetUsername, req.NewRole); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Role changed successfully",
	})
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	var req DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := services.DeleteUser(req.AdminUsername, req.TargetUsername); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
	})
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT username FROM users")
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			return
		}
		users = append(users, username)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"users": users,
	})
}
