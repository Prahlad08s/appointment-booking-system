package business

import (
	"context"
	"errors"
	"time"

	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"
	"appointment-booking-system/repositories"

	"go.uber.org/zap"
)

type CoachService struct {
	repo   *repositories.CoachRepository
	logger *zap.Logger
}

func NewCoachService(repo *repositories.CoachRepository, logger *zap.Logger) *CoachService {
	return &CoachService{repo: repo, logger: logger}
}

func (s *CoachService) CreateCoach(ctx context.Context, req models.CreateCoachRequest) (*models.Coach, error) {
	tz := req.Timezone
	if tz == "" {
		tz = constants.DefaultTimezone
	}

	if _, err := time.LoadLocation(tz); err != nil {
		return nil, errors.New(constants.ErrInvalidTimezone)
	}

	coach := &models.Coach{
		Name:     req.Name,
		Email:    req.Email,
		Timezone: tz,
	}

	if err := s.repo.Create(ctx, coach); err != nil {
		s.logger.Error("Failed to create coach", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}

	return coach, nil
}

func (s *CoachService) GetAllCoaches(ctx context.Context) ([]models.Coach, error) {
	coaches, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get coaches", zap.Error(err))
		return nil, errors.New(constants.ErrInternalServer)
	}
	return coaches, nil
}
