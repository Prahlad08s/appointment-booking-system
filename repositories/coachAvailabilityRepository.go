package repositories

import (
	"context"

	"appointment-booking-system/models"

	"gorm.io/gorm"
)

type CoachAvailabilityRepository struct {
	db *gorm.DB
}

func NewCoachAvailabilityRepository(db *gorm.DB) *CoachAvailabilityRepository {
	return &CoachAvailabilityRepository{db: db}
}

func (r *CoachAvailabilityRepository) Create(ctx context.Context, availability *models.CoachAvailability) error {
	return r.db.WithContext(ctx).Create(availability).Error
}

func (r *CoachAvailabilityRepository) GetByCoachID(ctx context.Context, coachID uint) ([]models.CoachAvailability, error) {
	var availabilities []models.CoachAvailability
	err := r.db.WithContext(ctx).
		Where("coach_id = ?", coachID).
		Order("day_of_week ASC, start_time ASC").
		Find(&availabilities).Error
	return availabilities, err
}

func (r *CoachAvailabilityRepository) GetByCoachAndDay(ctx context.Context, coachID uint, dayOfWeek int) ([]models.CoachAvailability, error) {
	var availabilities []models.CoachAvailability
	err := r.db.WithContext(ctx).
		Where("coach_id = ? AND day_of_week = ?", coachID, dayOfWeek).
		Order("start_time ASC").
		Find(&availabilities).Error
	return availabilities, err
}

func (r *CoachAvailabilityRepository) HasOverlap(ctx context.Context, coachID uint, dayOfWeek int, startTime, endTime string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.CoachAvailability{}).
		Where("coach_id = ? AND day_of_week = ? AND start_time < ? AND end_time > ?",
			coachID, dayOfWeek, endTime, startTime).
		Count(&count).Error
	return count > 0, err
}

func (r *CoachAvailabilityRepository) GetCoachByID(ctx context.Context, coachID uint) (*models.Coach, error) {
	var coach models.Coach
	err := r.db.WithContext(ctx).First(&coach, coachID).Error
	if err != nil {
		return nil, err
	}
	return &coach, nil
}
