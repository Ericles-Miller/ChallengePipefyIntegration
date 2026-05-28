// @title           ChallengePipefyIntegration API
// @version         1.0
// @description     API for ChallengePipefyIntegration
// @host            localhost:8080
// @BasePath        /
package main

import (
	"log"
	"os"

	_ "github.com/Ericles-Miller/ChallengePipefyIntegration/docs"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/api"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	server := api.NewServer()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := server.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
