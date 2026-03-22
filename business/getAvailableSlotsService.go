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

type GetAvailableSlotsService struct {
	availRepo *repositories.CoachAvailabilityRepository
	slotsRepo *repositories.SlotsRepository
	logger    *zap.Logger
}

func NewGetAvailableSlotsService(
	availRepo *repositories.CoachAvailabilityRepository,
	slotsRepo *repositories.SlotsRepository,
	logger *zap.Logger,
) *GetAvailableSlotsService {
	return &GetAvailableSlotsService{
		availRepo: availRepo,
		slotsRepo: slotsRepo,
		logger:    logger,
	}
}

func (s *GetAvailableSlotsService) GetSlots(ctx context.Context, coachID uint, dateStr, timezone string) ([]models.SlotResponse, error) {
	date, err := time.Parse(constants.DateFormat, dateStr)
	if err != nil {
		return nil, errors.New(constants.ErrInvalidDate)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if date.Before(today) {
		return nil, errors.New(constants.ErrDateInPast)
	}

	coach, err := s.availRepo.GetCoachByID(ctx, coachID)
	if err != nil {
		return nil, errors.New(constants.ErrCoachNotFound)
	}

	coachLocation, err := time.LoadLocation(coach.Timezone)
	if err != nil {
		s.logger.Error("Invalid coach timezone", zap.String("timezone", coach.Timezone), zap.Error(err))
		coachLocation = time.UTC
	}

	dayOfWeek := int(date.Weekday())

	availabilities, err := s.availRepo.GetByCoachAndDay(ctx, coachID, dayOfWeek)
	if err != nil {
		s.logger.Error("Failed to get availability", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}

	if len(availabilities) == 0 {
		return []models.SlotResponse{}, nil
	}

	var allSlots []commons.TimeSlot
	for _, avail := range availabilities {
		availStart, err := commons.ParseTimeString(avail.StartTime)
		if err != nil {
			s.logger.Error("Invalid start_time in availability", zap.Uint("id", avail.ID), zap.Error(err))
			continue
		}
		availEnd, err := commons.ParseTimeString(avail.EndTime)
		if err != nil {
			s.logger.Error("Invalid end_time in availability", zap.Uint("id", avail.ID), zap.Error(err))
			continue
		}
		slots := commons.GenerateSlots(date, availStart, availEnd, coachLocation)
		allSlots = append(allSlots, slots...)
	}

	// Get booked slots for the day (using a wide UTC window to cover timezone differences)
	dayStartUTC := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, coachLocation).UTC()
	dayEndUTC := dayStartUTC.Add(24 * time.Hour)

	bookedSlots, err := s.slotsRepo.GetBookedSlots(ctx, coachID, dayStartUTC, dayEndUTC)
	if err != nil {
		s.logger.Error("Failed to get booked slots", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}

	available := commons.FilterBookedSlots(allSlots, bookedSlots)

	userLocation, err := commons.ParseTimezone(timezone)
	if err != nil {
		return nil, errors.New(constants.ErrInvalidTimezone)
	}

	if userLocation != time.UTC {
		available = commons.ConvertSlotsToTimezone(available, userLocation)
	}

	response := make([]models.SlotResponse, len(available))
	for i, slot := range available {
		response[i] = models.SlotResponse{
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
		}
	}

	return response, nil
}
