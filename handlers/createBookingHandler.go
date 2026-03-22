package handlers

import (
	"net/http"

	"appointment-booking-system/business"
	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"github.com/gin-gonic/gin"
)

type CreateBookingHandler struct {
	service *business.CreateBookingService
}

func NewCreateBookingHandler(service *business.CreateBookingService) *CreateBookingHandler {
	return &CreateBookingHandler{service: service}
}

func (h *CreateBookingHandler) CreateBooking() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateBookingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(constants.ErrInvalidRequest))
			return
		}

		booking, err := h.service.CreateBooking(c.Request.Context(), req)
		if err != nil {
			switch err.Error() {
			case constants.ErrInvalidRequest, constants.ErrSlotNotAligned, constants.ErrSlotInPast, constants.ErrInvalidTimezone:
				c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
			case constants.ErrCoachNotFound:
				c.JSON(http.StatusNotFound, models.ErrorResponse(err.Error()))
			case constants.ErrSlotAlreadyBooked:
				c.JSON(http.StatusConflict, models.ErrorResponse(err.Error()))
			case constants.ErrSlotOutsideAvail:
				c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
			}
			return
		}

		c.JSON(http.StatusCreated, models.SuccessResponse("booking created successfully", booking))
	}
}
