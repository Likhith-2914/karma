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

	rows, err := db.DB.Query("SELECT id, title, description, status, story_points, due_date, project_id, user_id FROM tasks WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var dueDate sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.StoryPoints, &dueDate, &t.ProjectID, &t.UserID); err != nil {
			continue
		}
		if dueDate.Valid {
			t.DueDate = dueDate.String
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

	rows, err := db.DB.Query("SELECT id, title, description, status, story_points, due_date, project_id, user_id FROM tasks WHERE project_id = ? AND user_id = ?", projectID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var dueDate sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.StoryPoints, &dueDate, &t.ProjectID, &t.UserID); err != nil {
			continue
		}
		if dueDate.Valid {
			t.DueDate = dueDate.String
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

	result, err := db.DB.Exec(
		"INSERT INTO tasks (title, description, status, story_points, due_date, project_id, user_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		t.Title, t.Description, t.Status, t.StoryPoints, t.DueDate, t.ProjectID, userID,
	)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	t.ID = int(id)
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

	_, err := db.DB.Exec("UPDATE tasks SET status = ? WHERE id = ? AND user_id = ?", payload.Status, taskID, userID)
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

	_, err := db.DB.Exec(
		"UPDATE tasks SET title = ?, description = ?, status = ?, story_points = ?, due_date = ?, project_id = ? WHERE id = ? AND user_id = ?",
		t.Title, t.Description, t.Status, t.StoryPoints, t.DueDate, t.ProjectID, taskID, userID,
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

	_, err := db.DB.Exec("DELETE FROM tasks WHERE id = ? AND user_id = ?", taskID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted successfully"})
}
