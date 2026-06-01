package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"mi-api-go/internal/adapters/database"
	dbHttp "mi-api-go/internal/adapters/http"
	"mi-api-go/internal/core/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Lee desde variable de entorno, si no existe usa el valor local por defecto
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://root:secretpassword@localhost:5432/user_db?sslmode=disable"
	}

	jwtSecret := "mi_clave_secreta_super_segura_2026"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Error de conexión a Postgres: %v", err)
	}
	defer pool.Close()

	userRepo := database.NewPostgresRepository(pool)
	userService := services.NewUserService(userRepo, jwtSecret)
	userHandler := dbHttp.NewUserHandler(userService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Post("/auth/register", userHandler.Register)
	r.Post("/auth/login", userHandler.Login)

	r.Group(func(protected chi.Router) {
		protected.Use(dbHttp.JWTMiddleware(jwtSecret))
		protected.Get("/users/{id}", userHandler.GetUser)
		protected.Get("/users", userHandler.ListUsers)
		protected.Put("/users/{id}", userHandler.UpdateUser)
		protected.Delete("/users/{id}", userHandler.DeleteUser)
		protected.Post("/users/upload", userHandler.UploadFile)
	})

	log.Println("Servidor Go corriendo en http://localhost:8080 🚀")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error al arrancar: %v", err)
	}
}
