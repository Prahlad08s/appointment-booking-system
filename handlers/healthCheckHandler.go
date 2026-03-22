package handlers

import (
	"net/http"

	"appointment-booking-system/models"

	"github.com/gin-gonic/gin"
)

func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, models.SuccessResponse("service is healthy", nil))
	}
}
