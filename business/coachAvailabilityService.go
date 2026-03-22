package business

import (
	"context"
	"errors"

	"appointment-booking-system/commons"
	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"
	"appointment-booking-system/repositories"

	"go.uber.org/zap"
)

type CoachAvailabilityService struct {
	repo   *repositories.CoachAvailabilityRepository
	logger *zap.Logger
}

func NewCoachAvailabilityService(repo *repositories.CoachAvailabilityRepository, logger *zap.Logger) *CoachAvailabilityService {
	return &CoachAvailabilityService{repo: repo, logger: logger}
}

func (s *CoachAvailabilityService) SetAvailability(ctx context.Context, req models.SetAvailabilityRequest) (*models.CoachAvailability, error) {
	if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
		return nil, errors.New(constants.ErrInvalidDayOfWeek)
	}

	normalizedStart, err := commons.NormalizeTimeString(req.StartTime)
	if err != nil {
		return nil, errors.New(constants.ErrInvalidRequest)
	}

	normalizedEnd, err := commons.NormalizeTimeString(req.EndTime)
	if err != nil {
		return nil, errors.New(constants.ErrInvalidRequest)
	}

	startTime, _ := commons.ParseTimeString(normalizedStart)
	endTime, _ := commons.ParseTimeString(normalizedEnd)

	if !startTime.Before(endTime) {
		return nil, errors.New(constants.ErrInvalidTimeRange)
	}

	hasOverlap, err := s.repo.HasOverlap(ctx, req.CoachID, req.DayOfWeek, normalizedStart, normalizedEnd)
	if err != nil {
		s.logger.Error("Failed to check overlap", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}
	if hasOverlap {
		return nil, errors.New(constants.ErrOverlappingAvail)
	}

	availability := &models.CoachAvailability{
		CoachID:   req.CoachID,
		DayOfWeek: req.DayOfWeek,
		StartTime: normalizedStart,
		EndTime:   normalizedEnd,
	}

	if err := s.repo.Create(ctx, availability); err != nil {
		s.logger.Error("Failed to create availability", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}

	return availability, nil
}

func (s *CoachAvailabilityService) GetAvailability(ctx context.Context, coachID uint) ([]models.AvailabilityResponse, error) {
	availabilities, err := s.repo.GetByCoachID(ctx, coachID)
	if err != nil {
		s.logger.Error("Failed to get availability", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}

	response := make([]models.AvailabilityResponse, 0, len(availabilities))
	for _, a := range availabilities {
		response = append(response, models.AvailabilityResponse{
			ID:        a.ID,
			CoachID:   a.CoachID,
			DayOfWeek: a.DayOfWeek,
			StartTime: a.StartTime,
			EndTime:   a.EndTime,
		})
	}

	return response, nil
}
