// Command api is the entry point for the Hospital Middleware API service.
package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MoriMokata/hospital-middleware-system/internal/config"
	"github.com/MoriMokata/hospital-middleware-system/internal/handler"
	"github.com/MoriMokata/hospital-middleware-system/internal/repository"
	"github.com/MoriMokata/hospital-middleware-system/internal/service"
	"github.com/MoriMokata/hospital-middleware-system/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := migrations.Up(context.Background(), db); err != nil {
		log.Fatalf("migrate up: %v", err)
	}

	hospitalRepo := repository.NewPostgresHospitalRepository(db)
	staffRepo := repository.NewPostgresStaffRepository(db)

	staffService := service.NewStaffService(hospitalRepo, staffRepo, cfg.JWTSecret, cfg.JWTExpiry)
	staffHandler := handler.NewStaffHandler(staffService)

	router := gin.Default()
	router.GET("/health", handler.Health)
	router.POST("/staff/create", staffHandler.Create)
	router.POST("/staff/login", staffHandler.Login)

	log.Printf("hospital-middleware-system listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
