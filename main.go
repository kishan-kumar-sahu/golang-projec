// package main

// import (
// 	"log"
// 	"net/http"
// 	"time"

// 	"github.com/gorilla/mux"

// 	"golang_project/database"
// 	"golang_project/handlers"
// 	"golang_project/middleware"
// )

// func main() {
// 	db, err := database.InitDB("users.db")
// 	if err != nil {
// 		log.Fatalf("failed to initialize database: %v", err)
// 	}
// 	defer db.Close()

// 	router := mux.NewRouter()

// 	router.HandleFunc("/", handlers.HomeHandler).Methods(http.MethodGet)
// 	router.HandleFunc("/test", handlers.TestPageHandler).Methods(http.MethodGet)
// 	router.HandleFunc("/signup", handlers.SignupHandler(db)).Methods(http.MethodPost)
// 	router.HandleFunc("/login", handlers.LoginHandler(db)).Methods(http.MethodPost)

// 	protected := router.NewRoute().Subrouter()
// 	protected.Use(middleware.JWTMiddleware)
// 	protected.HandleFunc("/profile", handlers.ProfileHandler(db)).Methods(http.MethodGet)
// 	protected.HandleFunc("/users", handlers.UsersHandler(db)).Methods(http.MethodGet)

// 	server := &http.Server{
// 		Addr:         ":8081",
// 		Handler:      router,
// 		ReadTimeout:  15 * time.Second,
// 		WriteTimeout: 15 * time.Second,
// 	}

// 	log.Println("server listening on http://localhost:8081")
// 	if err := server.ListenAndServe(); err != nil {
// 		log.Fatalf("server error: %v", err)
// 	}
// }



package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"golang_project/database"
	"golang_project/handlers"
	"golang_project/middleware"
)

func main() {

	db, err := database.InitDB("users.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	router := mux.NewRouter()

	// Public Routes
	router.HandleFunc("/", handlers.HomeHandler).Methods(http.MethodGet)
	router.HandleFunc("/test", handlers.TestPageHandler).Methods(http.MethodGet)
	router.HandleFunc("/signup", handlers.SignupHandler(db)).Methods(http.MethodPost)
	router.HandleFunc("/login", handlers.LoginHandler(db)).Methods(http.MethodPost)

	// Protected Routes
	protected := router.NewRoute().Subrouter()
	protected.Use(middleware.JWTMiddleware)

	protected.HandleFunc("/profile", handlers.ProfileHandler(db)).Methods(http.MethodGet)
	protected.HandleFunc("/users", handlers.UsersHandler(db)).Methods(http.MethodGet)

	// Dynamic Port for Render
	port := os.Getenv("PORT")

	if port == "" {
		port = "8081"
	}

	// HTTP Server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("server listening on port", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}