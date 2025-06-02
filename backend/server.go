// server.go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"backend/config"
	"backend/controllers"
	"backend/database"
	authmw "backend/middleware"
	"backend/repositories"
	"backend/routes"
	"backend/services"
)

func main() {
	// ------------------------------------------------------------------
	// 1. Config
	// ------------------------------------------------------------------
	ctx := context.Background()
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// ------------------------------------------------------------------
	// 2. Database
	// ------------------------------------------------------------------
	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// ------------------------------------------------------------------
	// 3. Echo instance + global middleware
	// ------------------------------------------------------------------
	e := echo.New()
	// Set the log level to DEBUG to see more information
	e.Logger.SetLevel(log.DEBUG)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())
	e.Use(middleware.Secure())

	// 🔒 Auth0: validate every incoming Bearer token
	log.Printf("Registering Auth0 middleware...")
	e.Use(authmw.Auth0(cfg))
	log.Printf("Auth0 middleware registered")

	// ------------------------------------------------------------------
	// 4. Feature modules & routes
	// ------------------------------------------------------------------
	// Category
	categoryRepo := repositories.NewCategoryRepository(db.DB)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryController := controllers.NewCategoryController(categoryService)
	routes.SetupCategoryRoutes(e, categoryController)

	// Auth (login / callback / logout helpers)
	authCfg := services.AuthConfig{
		ClientID:     cfg.AUTH0.ClientID,
		ClientSecret: cfg.AUTH0.ClientSecret,
		Domain:       cfg.AUTH0.Domain,
		BaseURL:      cfg.GetBaseURL(),
	}
	authService, _ := services.NewAuthService(authCfg)
	authCtrl := controllers.NewAuthController(authService)
	routes.SetupAuthRoutes(e, authCtrl)

	// ------------------------------------------------------------------
	// 5. Start server
	// ------------------------------------------------------------------
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	if err := e.Start(addr); err != nil {
		log.Printf("Shutting down the server: %v", err)
		os.Exit(1)
	}
}
