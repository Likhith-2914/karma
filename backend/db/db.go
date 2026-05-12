package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error
	
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		// Fallback for local development if needed, though you should export DATABASE_URL locally
		dbUrl = "postgres://postgres:postgres@localhost:5432/karma?sslmode=disable"
		log.Println("DATABASE_URL not set, using default local postgres url")
	}

	DB, err = sql.Open("postgres", dbUrl)
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
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		theme TEXT DEFAULT 'dark'
	);`

	projectTable := `
	CREATE TABLE IF NOT EXISTS projects (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		user_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	taskTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL,
		story_points INTEGER DEFAULT 0,
		due_date TEXT,
		project_id INTEGER,
		user_id INTEGER,
		position INTEGER DEFAULT 0,
		FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE SET NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
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

	// Add position column to existing tables
	if _, err := DB.Exec("ALTER TABLE tasks ADD COLUMN IF NOT EXISTS position INTEGER DEFAULT 0;"); err != nil {
		log.Fatalf("Failed to add position column to tasks table: %v", err)
	}

	log.Println("Database tables initialized successfully")
}
