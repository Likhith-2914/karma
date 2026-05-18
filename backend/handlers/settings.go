package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"karma-backend/db"
	"karma-backend/models"
)

func GetSprintSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)

	var s models.UserSettings
	var startDate, endDate sql.NullString
	err := db.DB.QueryRow("SELECT sprint_start_date, sprint_end_date, target_points FROM user_settings WHERE user_id = $1", userID).
		Scan(&startDate, &endDate, &s.TargetPoints)

	if err == sql.ErrNoRows {
		// Default to current week if no sprint exists
		now := time.Now()
		offset := int(time.Monday - now.Weekday())
		if offset > 0 {
			offset -= 7
		}
		startOfWeek := time.Date(now.Year(), now.Month(), now.Day()+offset, 0, 0, 0, 0, now.Location())
		endOfWeek := startOfWeek.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		s = models.UserSettings{
			UserID:          userID,
			SprintStartDate: startOfWeek.Format(time.RFC3339),
			SprintEndDate:   endOfWeek.Format(time.RFC3339),
			TargetPoints:    60,
		}
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	} else {
		s.UserID = userID
		if startDate.Valid {
			s.SprintStartDate = startDate.String
		}
		if endDate.Valid {
			s.SprintEndDate = endDate.String
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func StartSprintHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)

	var payload struct {
		TargetPoints int `json:"target_points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	now := time.Now()
	// Calculate end of week (Sunday 23:59:59)
	daysUntilSunday := int(time.Sunday - now.Weekday())
	if daysUntilSunday < 0 {
		daysUntilSunday += 7 // If today is Sunday, daysUntilSunday is 0
	} else if daysUntilSunday == 0 && now.Weekday() != time.Sunday {
		daysUntilSunday = 7
	}

	endOfWeek := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSunday, 23, 59, 59, 0, now.Location())
	startDateStr := now.Format(time.RFC3339)
	endDateStr := endOfWeek.Format(time.RFC3339)

	// Upsert user_settings
	_, err := db.DB.Exec(`
		INSERT INTO user_settings (user_id, sprint_start_date, sprint_end_date, target_points) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET sprint_start_date = EXCLUDED.sprint_start_date, sprint_end_date = EXCLUDED.sprint_end_date, target_points = EXCLUDED.target_points
	`, userID, startDateStr, endDateStr, payload.TargetPoints)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"sprint_start_date": startDateStr,
		"sprint_end_date":   endDateStr,
		"message":           "Sprint started successfully",
	})
}
