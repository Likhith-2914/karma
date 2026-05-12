package models

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Name         string `json:"name"`
	Theme        string `json:"theme"`
}

type Project struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	UserID int    `json:"user_id"`
}

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	StoryPoints int    `json:"story_points"`
	DueDate     string `json:"due_date"`
	ProjectID   int    `json:"project_id"`
	UserID      int    `json:"user_id"`
	Position    int    `json:"position"`
}
