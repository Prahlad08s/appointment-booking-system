package handlers

import (
	"net/http"

	"appointment-booking-system/business"
	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"github.com/gin-gonic/gin"
)

type CoachHandler struct {
	service *business.CoachService
}

func NewCoachHandler(service *business.CoachService) *CoachHandler {
	return &CoachHandler{service: service}
}

func (h *CoachHandler) CreateCoach() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateCoachRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(constants.ErrInvalidRequest))
			return
		}

		coach, err := h.service.CreateCoach(c.Request.Context(), req)
		if err != nil {
			switch err.Error() {
			case constants.ErrInvalidTimezone:
				c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
			}
			return
		}

		c.JSON(http.StatusCreated, models.SuccessResponse("coach created successfully", coach))
	}
}

func (h *CoachHandler) GetCoaches() gin.HandlerFunc {
	return func(c *gin.Context) {
		coaches, err := h.service.GetAllCoaches(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse("coaches retrieved successfully", coaches))
	}
}
