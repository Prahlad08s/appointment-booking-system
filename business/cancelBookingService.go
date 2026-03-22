package business

import (
	"context"
	"errors"

	"appointment-booking-system/commons/constants"
	"appointment-booking-system/repositories"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CancelBookingService struct {
	bookingRepo *repositories.BookingRepository
	logger      *zap.Logger
}

func NewCancelBookingService(bookingRepo *repositories.BookingRepository, logger *zap.Logger) *CancelBookingService {
	return &CancelBookingService{bookingRepo: bookingRepo, logger: logger}
}

func (s *CancelBookingService) CancelBooking(ctx context.Context, bookingID, userID uint) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(constants.ErrBookingNotFound)
		}
		s.logger.Error("Failed to fetch booking", zap.Error(err))
		return errors.New(constants.ErrInternalServer)
	}

	if booking.UserID != userID {
		return errors.New(constants.ErrBookingNotOwned)
	}

	if booking.Status == constants.BookingStatusCancelled {
		return errors.New(constants.ErrAlreadyCancelled)
	}

	if err := s.bookingRepo.CancelBooking(ctx, bookingID); err != nil {
		s.logger.Error("Failed to cancel booking", zap.Error(err))
		return errors.New(constants.ErrInternalServer)
	}

	return nil
}
