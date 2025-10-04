package main

import (
	"log"

	"library-service/internal/infrastructure/app"
)

// @title Library Service API
// @version 2.0
// @description Library management system with clean architecture
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@library.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Create application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Run application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
