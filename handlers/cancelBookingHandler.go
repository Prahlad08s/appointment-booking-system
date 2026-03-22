package handlers

import (
	"net/http"
	"strconv"

	"appointment-booking-system/business"
	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"github.com/gin-gonic/gin"
)

type CancelBookingHandler struct {
	service *business.CancelBookingService
}

func NewCancelBookingHandler(service *business.CancelBookingService) *CancelBookingHandler {
	return &CancelBookingHandler{service: service}
}

func (h *CancelBookingHandler) CancelBooking() gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingIDStr := c.Param("id")
		bookingID, err := strconv.ParseUint(bookingIDStr, 10, 64)
		if err != nil || bookingID == 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(constants.ErrInvalidBookingID))
			return
		}

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

		if err := h.service.CancelBooking(c.Request.Context(), uint(bookingID), uint(userID)); err != nil {
			switch err.Error() {
			case constants.ErrBookingNotFound:
				c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
			case constants.ErrBookingNotOwned:
				c.JSON(http.StatusForbidden, models.ErrorResponse(err.Error()))
			case constants.ErrAlreadyCancelled:
				c.JSON(http.StatusConflict, models.ErrorResponse(err.Error()))
			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
			}
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse("booking cancelled successfully", nil))
	}
}
