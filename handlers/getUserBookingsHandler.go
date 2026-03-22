package handlers

import (
	"net/http"
	"strconv"

	"appointment-booking-system/business"
	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"github.com/gin-gonic/gin"
)

type GetUserBookingsHandler struct {
	service *business.GetUserBookingsService
}

func NewGetUserBookingsHandler(service *business.GetUserBookingsService) *GetUserBookingsHandler {
	return &GetUserBookingsHandler{service: service}
}

func (h *GetUserBookingsHandler) GetBookings() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(constants.ErrMissingUserID))
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil || userID == 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(constants.ErrInvalidUserID))
			return
		}

		bookings, err := h.service.GetBookings(c.Request.Context(), uint(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse("bookings retrieved successfully", bookings))
	}
}
