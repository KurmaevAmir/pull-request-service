package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/http/handlers"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/repositories"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/services"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/validators"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("env file not loaded: %v", err)
	}

	dsn := getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/pull_request_service?sslmode=disable")

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}

	teamRepo := repositories.NewPgTeamRepository(pool)
	teamValidator := validators.NewTeamValidator(teamRepo)
	teamService := services.NewTeamService(teamRepo, teamValidator)
	teamHandler := handlers.NewTeamHandler(teamService)

	userRepo := repositories.NewPgUserRepository(pool)
	userValidator := validators.NewUserValidator()
	userService := services.NewUserService(userRepo, userValidator)
	userHandler := handlers.NewUserHandler(userService)

	prRepo := repositories.NewPRRepository(pool)
	prValidator := validators.NewPRValidator(prRepo, userRepo)
	prService := services.NewPRService(prRepo, userRepo, prValidator)
	prHandler := handlers.NewPRHandler(prService)

	router := handlers.NewRouter(teamHandler, userHandler, prHandler)
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
