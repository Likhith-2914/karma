package main

import (
	"log"
	"net/http"
	"os"

	"karma-backend/db"
	"karma-backend/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// 1. Initialize the SQLite Database
	db.InitDB()

	// 2. Initialize the Router
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 3. Setup CORS so our React frontend can connect securely
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL}, // React URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// 4. Define our API Routes

	// Public routes (No authentication required)
	r.Post("/api/signup", handlers.SignupHandler)
	r.Post("/api/login", handlers.LoginHandler)

	// Protected routes (Require a valid JWT token)
	// We use chi's 'Route' grouping to apply the AuthMiddleware to a subset of routes.
	r.Route("/api", func(r chi.Router) {
		r.Use(handlers.AuthMiddleware)

		// User
		r.Patch("/user/theme", handlers.UpdateThemeHandler)

		// Projects
		r.Get("/projects", handlers.GetProjectsHandler)
		r.Post("/projects", handlers.CreateProjectHandler)
		r.Delete("/projects/{projectID}", handlers.DeleteProjectHandler)

		// All Tasks (for weekly view)
		r.Get("/tasks", handlers.GetAllTasksHandler)

		// Project Tasks
		r.Get("/projects/{projectID}/tasks", handlers.GetTasksHandler)
		r.Post("/tasks", handlers.CreateTaskHandler)
		r.Patch("/tasks/{taskID}/status", handlers.UpdateTaskStatusHandler) // PATCH is conventionally used for partial updates
		r.Put("/tasks/{taskID}", handlers.UpdateTaskHandler)
		r.Delete("/tasks/{taskID}", handlers.DeleteTaskHandler)
	})

	// 5. Start the Server
	log.Println("Starting Go backend server on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
