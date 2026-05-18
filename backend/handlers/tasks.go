package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"karma-backend/db"
	"karma-backend/models"

	"github.com/go-chi/chi/v5"
)

// GetAllTasksHandler returns all tasks across all projects (useful for Weekly Plan)
func GetAllTasksHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)

	rows, err := db.DB.Query("SELECT id, title, description, status, story_points, due_date, project_id, user_id, position, completed_at FROM tasks WHERE user_id = $1 ORDER BY position ASC, id ASC", userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var dueDate sql.NullString
		var projectID sql.NullInt64
		var completedAt sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.StoryPoints, &dueDate, &projectID, &t.UserID, &t.Position, &completedAt); err != nil {
			continue
		}
		if dueDate.Valid {
			t.DueDate = dueDate.String
		}
		if projectID.Valid {
			t.ProjectID = int(projectID.Int64)
		} else {
			t.ProjectID = 0
		}
		if completedAt.Valid {
			t.CompletedAt = completedAt.String
		}
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTasksHandler returns all tasks for a specific project
func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)
	projectID := chi.URLParam(r, "projectID")

	rows, err := db.DB.Query("SELECT id, title, description, status, story_points, due_date, project_id, user_id, position, completed_at FROM tasks WHERE project_id = $1 AND user_id = $2 ORDER BY position ASC, id ASC", projectID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var dueDate sql.NullString
		var projectID sql.NullInt64
		var completedAt sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.StoryPoints, &dueDate, &projectID, &t.UserID, &t.Position, &completedAt); err != nil {
			continue
		}
		if dueDate.Valid {
			t.DueDate = dueDate.String
		}
		if projectID.Valid {
			t.ProjectID = int(projectID.Int64)
		} else {
			t.ProjectID = 0
		}
		if completedAt.Valid {
			t.CompletedAt = completedAt.String
		}
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// CreateTaskHandler creates a new task
func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)

	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if t.Status == "" {
		t.Status = "TODO"
	}
	if t.DueDate == "" {
		// Default to today's date if not provided
		t.DueDate = time.Now().Format("2006-01-02")
	}

	var dbProjectID interface{} = t.ProjectID
	if t.ProjectID == 0 {
		dbProjectID = nil
	}

	var maxPos int
	err := db.DB.QueryRow("SELECT COALESCE(MAX(position), -1) FROM tasks WHERE user_id = $1 AND status = $2", userID, t.Status).Scan(&maxPos)
	if err == nil {
		t.Position = maxPos + 1
	}

	var completedAt interface{} = nil
	if t.Status == "COMPLETED" {
		completedAt = time.Now().Format(time.RFC3339)
	}

	var id int
	err = db.DB.QueryRow(
		"INSERT INTO tasks (title, description, status, story_points, due_date, project_id, user_id, position, completed_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		t.Title, t.Description, t.Status, t.StoryPoints, t.DueDate, dbProjectID, userID, t.Position, completedAt,
	).Scan(&id)
	
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	t.ID = id
	t.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// UpdateTaskStatusHandler updates only the status of a task
func UpdateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)
	taskID := chi.URLParam(r, "taskID")

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var completedAt interface{} = nil
	if payload.Status == "COMPLETED" {
		completedAt = time.Now().Format(time.RFC3339)
	}

	_, err := db.DB.Exec("UPDATE tasks SET status = $1, completed_at = $2 WHERE id = $3 AND user_id = $4", payload.Status, completedAt, taskID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Status updated successfully"})
}

// UpdateTaskHandler updates all editable fields of a task
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)
	taskID := chi.URLParam(r, "taskID")

	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if t.Status == "" {
		t.Status = "TODO"
	}
	if t.DueDate == "" {
		t.DueDate = time.Now().Format("2006-01-02")
	}

	var dbProjectID interface{} = t.ProjectID
	if t.ProjectID == 0 {
		dbProjectID = nil
	}

	var completedAt interface{} = nil
	if t.Status == "COMPLETED" {
		completedAt = time.Now().Format(time.RFC3339)
	}

	_, err := db.DB.Exec(
		"UPDATE tasks SET title = $1, description = $2, status = $3, story_points = $4, due_date = $5, project_id = $6, position = $7, completed_at = COALESCE($8, completed_at) WHERE id = $9 AND user_id = $10",
		t.Title, t.Description, t.Status, t.StoryPoints, t.DueDate, dbProjectID, t.Position, completedAt, taskID, userID,
	)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task updated successfully"})
}

// DeleteTaskHandler deletes a task
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)
	taskID := chi.URLParam(r, "taskID")

	_, err := db.DB.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", taskID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted successfully"})
}

// ReorderTasksHandler updates the status and position of multiple tasks
func ReorderTasksHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)

	var payload struct {
		Tasks []struct {
			ID       int    `json:"id"`
			Status   string `json:"status"`
			Position int    `json:"position"`
		} `json:"tasks"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	stmt, err := tx.Prepare(`
		UPDATE tasks 
		SET status = $1, 
			position = $2,
			completed_at = CASE WHEN $1 = 'COMPLETED' AND status != 'COMPLETED' THEN $3 
								WHEN $1 != 'COMPLETED' THEN NULL 
								ELSE completed_at END
		WHERE id = $4 AND user_id = $5
	`)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	nowStr := time.Now().Format(time.RFC3339)

	for _, task := range payload.Tasks {
		if _, err := stmt.Exec(task.Status, task.Position, nowStr, task.ID, userID); err != nil {
			tx.Rollback()
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tasks reordered successfully"})
}
