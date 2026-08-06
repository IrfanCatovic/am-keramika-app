package handlers

import (
	"net/http"

	"am-keramika-backend/database"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	status := "ok"
	dbStatus := "ok"
	code := http.StatusOK

	if err := database.Ping(); err != nil {
		status = "unavailable"
		dbStatus = "unavailable"
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, gin.H{
		"status":   status,
		"database": dbStatus,
	})
}
