package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health reports basic service liveness. Used as the docker-compose / nginx
// upstream health check and as the smoke-test target for Task 02's
// acceptance criteria (GET /health returns 200 via nginx).
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
