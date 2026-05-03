package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"karma-backend/db"
	"karma-backend/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT Secret Key. 
// In a real production application, this should be stored in an Environment Variable, not hardcoded!
var jwtSecret = []byte("karma-super-secret-key")

// Credentials is used to parse the login/signup JSON payload sent from React.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"` // omitempty: Only used for Signup, ignored for Login
}

// SignupHandler handles new user registration
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse JSON body coming from the frontend
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 2. Hash the password before storing it (CRITICAL FOR SECURITY)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// 3. Insert into Database
	_, err = db.DB.Exec("INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)", creds.Email, string(hashedPassword), creds.Name)
	if err != nil {
		// Usually fails here if the email UNIQUE constraint is violated
		http.Error(w, "Email already exists or database error", http.StatusConflict)
		return
	}

	// 4. Respond with success status (201 Created)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

// LoginHandler validates credentials and returns a JWT token
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse JSON body
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 2. Query the user by email from the DB
	var user models.User
	var theme sql.NullString
	row := db.DB.QueryRow("SELECT id, email, password_hash, name, theme FROM users WHERE email = $1", creds.Email)
	
	// Scan copies the columns from the matched row into our user struct variables
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &theme)
	if theme.Valid {
		user.Theme = theme.String
	} else {
		user.Theme = "dark"
	}

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// 3. Compare the stored hashed password with the provided plaintext password
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// 4. Create JWT Token (Used to keep the user logged in without sending password every time)
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	claims := &jwt.RegisteredClaims{
		Subject:   strconv.Itoa(user.ID), // We store the User ID in the token's subject
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign the token with our secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// 5. Send the token and user details back to the React app
	// The React app will save this token in localStorage and send it in the header of future requests.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"theme": user.Theme,
		},
	})
}

// UpdateThemeHandler updates the user's theme preference
func UpdateThemeHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, _ := strconv.Atoi(userIDStr)

	var payload struct {
		Theme string `json:"theme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("UPDATE users SET theme = $1 WHERE id = $2", payload.Theme, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Theme updated successfully"})
}
