package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"karma-backend/db"
	"karma-backend/models"

	"github.com/go-chi/chi/v5"
)

// GetProjectsHandler returns all projects for the logged-in user
func GetProjectsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the userID string from context (set by AuthMiddleware)
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)

	// Query the database
	rows, err := db.DB.Query("SELECT id, name, user_id FROM projects WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close() // Always defer closing the rows to prevent memory leaks

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.UserID); err != nil {
			continue
		}
		projects = append(projects, p)
	}

	// If projects is nil (user has 0 projects), return an empty array [] instead of null
	if projects == nil {
		projects = []models.Project{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// CreateProjectHandler creates a new project
func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)

	var p models.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute the INSERT statement and get the returned ID
	var id int
	err := db.DB.QueryRow("INSERT INTO projects (name, user_id) VALUES ($1, $2) RETURNING id", p.Name, userID).Scan(&id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	p.ID = id
	p.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(p)
}

// DeleteProjectHandler deletes a project and all its associated tasks
func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)
	projectID := chi.URLParam(r, "projectID")

	// Start a transaction since we are deleting tasks and then the project
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Delete tasks associated with the project
	_, err = tx.Exec("DELETE FROM tasks WHERE project_id = $1 AND user_id = $2", projectID, userID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete associated tasks", http.StatusInternalServerError)
		return
	}

	// Delete the project itself
	_, err = tx.Exec("DELETE FROM projects WHERE id = $1 AND user_id = $2", projectID, userID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Project and associated tasks deleted successfully"})
}
