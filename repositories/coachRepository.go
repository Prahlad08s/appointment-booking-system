package repositories

import (
	"context"

	"appointment-booking-system/models"

	"gorm.io/gorm"
)

type CoachRepository struct {
	db *gorm.DB
}

func NewCoachRepository(db *gorm.DB) *CoachRepository {
	return &CoachRepository{db: db}
}

func (r *CoachRepository) Create(ctx context.Context, coach *models.Coach) error {
	return r.db.WithContext(ctx).Create(coach).Error
}

func (r *CoachRepository) GetAll(ctx context.Context) ([]models.Coach, error) {
	var coaches []models.Coach
	err := r.db.WithContext(ctx).Order("id ASC").Find(&coaches).Error
	return coaches, err
}

func (r *CoachRepository) GetByID(ctx context.Context, id uint) (*models.Coach, error) {
	var coach models.Coach
	err := r.db.WithContext(ctx).First(&coach, id).Error
	if err != nil {
		return nil, err
	}
	return &coach, nil
}
