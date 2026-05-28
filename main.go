// @title           ChallengePipefyIntegration API
// @version         1.0
// @description     API for ChallengePipefyIntegration
// @host            localhost:8080
// @BasePath        /
package main

import (
	"context"
	"log"
	"os"

	"github.com/Ericles-Miller/ChallengePipefyIntegration/api"
	_ "github.com/Ericles-Miller/ChallengePipefyIntegration/docs"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	ctx := context.Background()

	pool, err := database.ConnectDB(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(pool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	server := api.NewServer(pool)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("PORT environment variable not set")
	}

	if err := server.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
