# Golang Backend Service

A small Go backend service with authentication, JWT middleware, user roles, and protected endpoints.

## APIs

- `POST /signup` - Create a new user account
- `POST /login` - Authenticate and receive JWT
- `GET /profile` - View the current user's profile (requires JWT)
- `GET /users` - List all users (requires JWT and Admin role)

## Setup

1. Install Go 1.24+.
2. Run:

```bash
cd "c:\Users\A S\OneDrive\Desktop\golang_project"
go mod tidy
go run .

```

3. Open `https://golang-projec.onrender.com/test`

## Notes

- Uses SQLite database file `users.db`.
- JWT secret is configured in `utils.go`; replace it with a secure secret before production.
- Passwords are hashed using `bcrypt`.
