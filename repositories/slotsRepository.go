package repositories

import (
	"context"
	"time"

	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"gorm.io/gorm"
)

type SlotsRepository struct {
	db *gorm.DB
}

func NewSlotsRepository(db *gorm.DB) *SlotsRepository {
	return &SlotsRepository{db: db}
}

// GetBookedSlots returns the start times of all active bookings for a coach on a given date range.
func (r *SlotsRepository) GetBookedSlots(ctx context.Context, coachID uint, dayStart, dayEnd time.Time) (map[time.Time]bool, error) {
	var bookings []models.Booking
	err := r.db.WithContext(ctx).
		Where("coach_id = ? AND start_time >= ? AND start_time < ? AND status = ?",
			coachID, dayStart, dayEnd, constants.BookingStatusBooked).
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}

	booked := make(map[time.Time]bool)
	for _, b := range bookings {
		booked[b.StartTime.UTC()] = true
	}
	return booked, nil
}
