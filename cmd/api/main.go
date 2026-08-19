// Command api is the entry point for the Hospital Middleware API service.
//
// Task 02 wires up the minimal HTTP server and a /health endpoint so the
// docker-compose stack (nginx + api + postgres) can be brought up and
// smoke-tested end to end. Routing for the real business endpoints
// (/staff/create, /staff/login, /patient/search) is added in later tasks.
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/MoriMokata/hospital-middleware-system/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.Default()
	router.GET("/health", handler.Health)

	log.Printf("hospital-middleware-system listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
