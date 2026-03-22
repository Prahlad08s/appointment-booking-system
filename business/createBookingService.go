package business

import (
	"context"
	"errors"
	"time"

	"appointment-booking-system/commons"
	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"
	"appointment-booking-system/repositories"

	"go.uber.org/zap"
)

type CreateBookingService struct {
	bookingRepo *repositories.BookingRepository
	availRepo   *repositories.CoachAvailabilityRepository
	logger      *zap.Logger
}

func NewCreateBookingService(
	bookingRepo *repositories.BookingRepository,
	availRepo *repositories.CoachAvailabilityRepository,
	logger *zap.Logger,
) *CreateBookingService {
	return &CreateBookingService{
		bookingRepo: bookingRepo,
		availRepo:   availRepo,
		logger:      logger,
	}
}

func (s *CreateBookingService) CreateBooking(ctx context.Context, req models.CreateBookingRequest) (*models.BookingResponse, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, errors.New(constants.ErrInvalidRequest)
	}

	if !commons.IsSlotAligned(startTime) {
		return nil, errors.New(constants.ErrSlotNotAligned)
	}

	if startTime.Before(time.Now().UTC()) {
		return nil, errors.New(constants.ErrSlotInPast)
	}

	userTZ := req.UserTimezone
	if userTZ == "" {
		userTZ = constants.DefaultTimezone
	}
	if _, err := time.LoadLocation(userTZ); err != nil {
		return nil, errors.New(constants.ErrInvalidTimezone)
	}

	coach, err := s.availRepo.GetCoachByID(ctx, req.CoachID)
	if err != nil {
		return nil, errors.New(constants.ErrCoachNotFound)
	}

	coachLocation, err := time.LoadLocation(coach.Timezone)
	if err != nil {
		coachLocation = time.UTC
	}

	// Validate that slot falls within coach availability
	coachLocalTime := startTime.In(coachLocation)
	dayOfWeek := int(coachLocalTime.Weekday())

	availabilities, err := s.availRepo.GetByCoachAndDay(ctx, req.CoachID, dayOfWeek)
	if err != nil {
		s.logger.Error("Failed to get coach availability", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}

	if !isWithinAvailability(coachLocalTime, availabilities) {
		return nil, errors.New(constants.ErrSlotOutsideAvail)
	}

	endTime := startTime.Add(time.Duration(constants.SlotDurationMinutes) * time.Minute)

	booking := &models.Booking{
		UserID:       req.UserID,
		CoachID:      req.CoachID,
		StartTime:    startTime.UTC(),
		EndTime:      endTime.UTC(),
		Status:       constants.BookingStatusBooked,
		UserTimezone: userTZ,
	}

	if err := s.bookingRepo.CreateBooking(ctx, booking); err != nil {
		if err.Error() == constants.ErrSlotAlreadyBooked {
			return nil, err
		}
		s.logger.Error("Failed to create booking", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}

	return &models.BookingResponse{
		ID:           booking.ID,
		UserID:       booking.UserID,
		CoachID:      booking.CoachID,
		StartTime:    booking.StartTime,
		EndTime:      booking.EndTime,
		Status:       booking.Status,
		UserTimezone: booking.UserTimezone,
		CreatedAt:    booking.CreatedAt,
	}, nil
}

func isWithinAvailability(t time.Time, availabilities []models.CoachAvailability) bool {
	slotTime := t.Format(constants.TimeFormat)
	slotEnd := t.Add(time.Duration(constants.SlotDurationMinutes) * time.Minute).Format(constants.TimeFormat)

	for _, avail := range availabilities {
		if slotTime >= avail.StartTime && slotEnd <= avail.EndTime {
			return true
		}
	}
	return false
}
