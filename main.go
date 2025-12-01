package main

import "library-service/internal/app"

// @title Library Service API
// @version 1.0
// @description Library management service with authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@library.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:80
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	app.Run()
}
