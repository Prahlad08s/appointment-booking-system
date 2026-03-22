package handlers

import (
	"net/http"
	"strconv"

	"appointment-booking-system/business"
	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"github.com/gin-gonic/gin"
)

type GetAvailableSlotsHandler struct {
	service *business.GetAvailableSlotsService
}

func NewGetAvailableSlotsHandler(service *business.GetAvailableSlotsService) *GetAvailableSlotsHandler {
	return &GetAvailableSlotsHandler{service: service}
}

func (h *GetAvailableSlotsHandler) GetSlots() gin.HandlerFunc {
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

		dateStr := c.Query("date")
		if dateStr == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(constants.ErrMissingDate))
			return
		}

		timezone := c.DefaultQuery("timezone", constants.DefaultTimezone)

		slots, err := h.service.GetSlots(c.Request.Context(), uint(coachID), dateStr, timezone)
		if err != nil {
			switch err.Error() {
			case constants.ErrInvalidDate, constants.ErrDateInPast, constants.ErrInvalidTimezone:
				c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
			case constants.ErrCoachNotFound:
				c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
			}
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse("available slots retrieved", slots))
	}
}
