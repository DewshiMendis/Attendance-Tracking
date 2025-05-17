package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite" //  with a blank identifier, correct import
)

var (
	DB *sql.DB
)

func InitDB() {
	var err error
	// Open SQLite database (this creates the file if this doesn't exist)
	DB, err = sql.Open("sqlite", "./attendance.db") // SQLite driver for pure Go implementation
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	// Creating tables if they don't exist
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        username TEXT PRIMARY KEY,
        password_hash TEXT,
        secret TEXT,
        role TEXT DEFAULT 'user'
    );
    CREATE TABLE IF NOT EXISTS attendance (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT,
        timestamp DATETIME,
        status TEXT
    );
    CREATE TABLE IF NOT EXISTS audit_logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT,
        event_type TEXT,
        success BOOLEAN,
        event_time DATETIME,
        ip_address TEXT
    );
    `
	_, err = DB.Exec(createTableQuery)
	if err != nil {
		log.Fatal("Error creating tables: ", err)
	}

	fmt.Println("Database setup complete!")
}

func Close() {
	DB.Close()
}
