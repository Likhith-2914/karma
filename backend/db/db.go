package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./karma.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	createTables()
}

func createTables() {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		theme TEXT DEFAULT 'dark'
	);`

	projectTable := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		user_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	taskTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL,
		story_points INTEGER DEFAULT 0,
		due_date TEXT,
		project_id INTEGER,
		user_id INTEGER,
		FOREIGN KEY(project_id) REFERENCES projects(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	if _, err := DB.Exec(userTable); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	if _, err := DB.Exec(projectTable); err != nil {
		log.Fatalf("Failed to create projects table: %v", err)
	}
	if _, err := DB.Exec(taskTable); err != nil {
		log.Fatalf("Failed to create tasks table: %v", err)
	}

	// Safety check: if the table was already created in a previous run without due_date,
	// this will add it. We ignore the error because it fails safely if the column exists.
	DB.Exec("ALTER TABLE tasks ADD COLUMN due_date TEXT")
	DB.Exec("ALTER TABLE users ADD COLUMN theme TEXT DEFAULT 'dark'")

	log.Println("Database tables initialized successfully")
}
