package handlers

import (
	"net/http"
	"strconv"

	"appointment-booking-system/business"
	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"github.com/gin-gonic/gin"
)

type CoachAvailabilityHandler struct {
	service *business.CoachAvailabilityService
}

func NewCoachAvailabilityHandler(service *business.CoachAvailabilityService) *CoachAvailabilityHandler {
	return &CoachAvailabilityHandler{service: service}
}

func (h *CoachAvailabilityHandler) SetAvailability() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SetAvailabilityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(constants.ErrInvalidRequest))
			return
		}

		availability, err := h.service.SetAvailability(c.Request.Context(), req)
		if err != nil {
			switch err.Error() {
			case constants.ErrInvalidDayOfWeek, constants.ErrInvalidTimeRange, constants.ErrInvalidRequest:
				c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
			case constants.ErrOverlappingAvail:
				c.JSON(http.StatusConflict, models.ErrorResponse(err.Error()))
			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
			}
			return
		}

		c.JSON(http.StatusCreated, models.SuccessResponse("availability set successfully", availability))
	}
}

func (h *CoachAvailabilityHandler) GetAvailability() gin.HandlerFunc {
	return func(c *gin.Context) {
		coachIDStr := c.Query("coach_id")
		if coachIDStr == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(constants.ErrMissingCoachID))
			return
		}

		coachID, err := strconv.ParseUint(coachIDStr, 10, 64)
		if err != nil || coachID == 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(constants.ErrInvalidCoachID))
			return
		}

		availabilities, err := h.service.GetAvailability(c.Request.Context(), uint(coachID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse("availability retrieved successfully", availabilities))
	}
}
