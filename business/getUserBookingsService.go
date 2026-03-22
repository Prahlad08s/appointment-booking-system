package business

import (
	"context"
	"errors"

	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"
	"appointment-booking-system/repositories"

	"go.uber.org/zap"
)

type GetUserBookingsService struct {
	bookingRepo *repositories.BookingRepository
	logger      *zap.Logger
}

func NewGetUserBookingsService(bookingRepo *repositories.BookingRepository, logger *zap.Logger) *GetUserBookingsService {
	return &GetUserBookingsService{bookingRepo: bookingRepo, logger: logger}
}

func (s *GetUserBookingsService) GetBookings(ctx context.Context, userID uint) ([]models.BookingResponse, error) {
	bookings, err := s.bookingRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user bookings", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}

	response := make([]models.BookingResponse, len(bookings))
	for i, b := range bookings {
		response[i] = models.BookingResponse{
			ID:           b.ID,
			UserID:       b.UserID,
			CoachID:      b.CoachID,
			StartTime:    b.StartTime,
			EndTime:      b.EndTime,
			Status:       b.Status,
			UserTimezone: b.UserTimezone,
			CreatedAt:    b.CreatedAt,
		}
	}

	return response, nil
}
