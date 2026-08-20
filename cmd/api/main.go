// Command api is the entry point for the Hospital Middleware API service.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MoriMokata/hospital-middleware-system/internal/config"
	"github.com/MoriMokata/hospital-middleware-system/internal/handler"
	"github.com/MoriMokata/hospital-middleware-system/internal/his"
	"github.com/MoriMokata/hospital-middleware-system/internal/middleware"
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
	patientRepo := repository.NewPostgresPatientRepository(db)

	staffService := service.NewStaffService(hospitalRepo, staffRepo, cfg.JWTSecret, cfg.JWTExpiry)
	staffHandler := handler.NewStaffHandler(staffService)

	// One HISClientFactory per hospitals.his_adapter_type value. Onboarding
	// a new hospital means registering a new factory here, not touching
	// the service/handler layers (see docs/architecture.md).
	hisFactories := map[string]service.HISClientFactory{
		"hospital_a": func(baseURL string) his.HISClient {
			return his.NewHospitalAClient(baseURL, &http.Client{Timeout: 10 * time.Second})
		},
	}
	patientService := service.NewPatientService(hospitalRepo, patientRepo, hisFactories)
	patientHandler := handler.NewPatientHandler(patientService)

	router := gin.Default()
	router.GET("/health", handler.Health)
	router.POST("/staff/create", staffHandler.Create)
	router.POST("/staff/login", staffHandler.Login)
	router.POST("/patient/search", middleware.Auth(cfg.JWTSecret), patientHandler.Search)

	log.Printf("hospital-middleware-system listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
