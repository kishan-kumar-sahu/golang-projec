package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang_project/models"
	"golang_project/middleware"
	"golang_project/utils"
)

func SignupHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		req.Role = strings.ToLower(strings.TrimSpace(req.Role))
		if req.Username == "" || req.Password == "" {
			respondError(w, http.StatusBadRequest, "username and password are required")
			return
		}

		if req.Role == "" {
			req.Role = "user"
		}
		if req.Role != "user" && req.Role != "admin" {
			respondError(w, http.StatusBadRequest, "role must be either 'user' or 'admin'")
			return
		}

		passwordHash, err := utils.HashPassword(req.Password)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to hash password")
			return
		}

		stmt, err := db.Prepare("INSERT INTO users(username, password_hash, role, created_at) VALUES (?, ?, ?, ?)")
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to prepare statement")
			return
		}
		defer stmt.Close()

		result, err := stmt.Exec(req.Username, passwordHash, req.Role, time.Now().UTC())
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				respondError(w, http.StatusConflict, "username already exists")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to save user")
			return
		}

		id, _ := result.LastInsertId()
		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":       id,
			"username": req.Username,
			"role":     req.Role,
		})
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}

		user, err := findUserByUsername(db, req.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				respondError(w, http.StatusUnauthorized, "invalid credentials")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to query user")
			return
		}

		if err := utils.ComparePassword(user.PasswordHash, req.Password); err != nil {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		token, err := utils.CreateJWT(user)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create token")
			return
		}

		respondJSON(w, http.StatusOK, models.AuthResponse{Token: token})
	}
}

func ProfileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current := middleware.GetCurrentUser(r)
		if current == nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		user, err := findUserByID(db, current.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				respondError(w, http.StatusNotFound, "user not found")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to query profile")
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"role":      user.Role,
			"createdAt": user.CreatedAt,
		})
	}
}

func UsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current := middleware.GetCurrentUser(r)
		if current == nil || current.Role != "admin" {
			respondError(w, http.StatusForbidden, "admin access required")
			return
		}

		rows, err := db.Query("SELECT id, username, role, created_at FROM users ORDER BY id")
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to read users")
			return
		}
		defer rows.Close()

		var users []map[string]interface{}
		for rows.Next() {
			var u models.User
			if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt); err != nil {
				respondError(w, http.StatusInternalServerError, "failed to scan users")
				return
			}
			users = append(users, map[string]interface{}{
				"id":        u.ID,
				"username":  u.Username,
				"role":      u.Role,
				"createdAt": u.CreatedAt,
			})
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"users": users})
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Welcome to the User Management API",
		"endpoints": map[string]interface{}{
			"signup":  "POST /signup",
			"login":   "POST /login",
			"profile": "GET /profile (requires JWT token)",
			"users":   "GET /users (requires JWT token with admin role)",
		},
	})
}

func TestPageHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>User Management API Test</title>
	<style>
		body { font-family: Arial; margin: 20px; }
		.container { max-width: 1000px; margin: 0 auto; }
		.section { border: 1px solid #ccc; padding: 20px; margin: 20px 0; }
		input, textarea { width: 100%; padding: 8px; margin: 5px 0; }
		button { padding: 10px 20px; margin: 10px 0; cursor: pointer; }
		.response { background: #f0f0f0; padding: 10px; margin: 10px 0; border-radius: 4px; white-space: pre-wrap; }
		.error { color: red; }
		.success { color: green; }
	</style>
</head>
<body>
	<div class="container">
		<h1>User Management API Test</h1>

		<div class="section">
			<h2>Signup</h2>
			<input type="text" id="signup-username" placeholder="Username">
			<input type="password" id="signup-password" placeholder="Password">
			<select id="signup-role">
				<option value="user">User</option>
				<option value="admin">Admin</option>
			</select>
			<button onclick="signup()">Sign Up</button>
			<div id="signup-response" class="response" style="display:none;"></div>
		</div>

		<div class="section">
			<h2>Login</h2>
			<input type="text" id="login-username" placeholder="Username">
			<input type="password" id="login-password" placeholder="Password">
			<button onclick="login()">Login</button>
			<div id="login-response" class="response" style="display:none;"></div>
		</div>

		<div class="section">
			<h2>Profile</h2>
			<input type="text" id="profile-token" placeholder="JWT Token (from login)">
			<button onclick="getProfile()">Get Profile</button>
			<div id="profile-response" class="response" style="display:none;"></div>
		</div>

		<div class="section">
			<h2>Users (Admin Only)</h2>
			<input type="text" id="users-token" placeholder="JWT Token (from login with admin role)">
			<button onclick="getUsers()">Get All Users</button>
			<div id="users-response" class="response" style="display:none;"></div>
		</div>
	</div>

	<script>
		async function signup() {
			const username = document.getElementById('signup-username').value;
			const password = document.getElementById('signup-password').value;
			const role = document.getElementById('signup-role').value;

			try {
				const response = await fetch('/signup', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ username, password, role })
				});
				const data = await response.json();
				const div = document.getElementById('signup-response');
				div.innerHTML = JSON.stringify(data, null, 2);
				div.className = 'response ' + (response.ok ? 'success' : 'error');
				div.style.display = 'block';
			} catch (e) {
				const div = document.getElementById('signup-response');
				div.innerHTML = 'Error: ' + e.message;
				div.className = 'response error';
				div.style.display = 'block';
			}
		}

		async function login() {
			const username = document.getElementById('login-username').value;
			const password = document.getElementById('login-password').value;

			try {
				const response = await fetch('/login', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ username, password })
				});
				const data = await response.json();
				const div = document.getElementById('login-response');
				div.innerHTML = JSON.stringify(data, null, 2);
				div.className = 'response ' + (response.ok ? 'success' : 'error');
				div.style.display = 'block';

				// Auto-fill token in profile and users fields
				if (response.ok && data.token) {
					document.getElementById('profile-token').value = data.token;
					document.getElementById('users-token').value = data.token;
				}
			} catch (e) {
				const div = document.getElementById('login-response');
				div.innerHTML = 'Error: ' + e.message;
				div.className = 'response error';
				div.style.display = 'block';
			}
		}

		async function getProfile() {
			const token = document.getElementById('profile-token').value;
			try {
				const response = await fetch('/profile', {
					method: 'GET',
					headers: { 'Authorization': 'Bearer ' + token }
				});
				const data = await response.json();
				const div = document.getElementById('profile-response');
				div.innerHTML = JSON.stringify(data, null, 2);
				div.className = 'response ' + (response.ok ? 'success' : 'error');
				div.style.display = 'block';
			} catch (e) {
				const div = document.getElementById('profile-response');
				div.innerHTML = 'Error: ' + e.message;
				div.className = 'response error';
				div.style.display = 'block';
			}
		}

		async function getUsers() {
			const token = document.getElementById('users-token').value;
			try {
				const response = await fetch('/users', {
					method: 'GET',
					headers: { 'Authorization': 'Bearer ' + token }
				});
				const data = await response.json();
				const div = document.getElementById('users-response');
				div.innerHTML = JSON.stringify(data, null, 2);
				div.className = 'response ' + (response.ok ? 'success' : 'error');
				div.style.display = 'block';
			} catch (e) {
				const div = document.getElementById('users-response');
				div.innerHTML = 'Error: ' + e.message;
				div.className = 'response error';
				div.style.display = 'block';
			}
		}
	</script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func findUserByUsername(db *sql.DB, username string) (*models.User, error) {
	var user models.User
	row := db.QueryRow("SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?", username)
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func findUserByID(db *sql.DB, id int64) (*models.User, error) {
	var user models.User
	row := db.QueryRow("SELECT id, username, password_hash, role, created_at FROM users WHERE id = ?", id)
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}