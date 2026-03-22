package repositories

import (
	"context"
	"errors"

	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// CreateBooking uses a transaction with row-level locking to prevent double booking.
func (r *BookingRepository) CreateBooking(ctx context.Context, booking *models.Booking) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.Booking
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("coach_id = ? AND start_time = ? AND status = ?",
				booking.CoachID, booking.StartTime, constants.BookingStatusBooked).
			First(&existing)

		if result.RowsAffected > 0 {
			return errors.New(constants.ErrSlotAlreadyBooked)
		}

		return tx.Create(booking).Error
	})
}

func (r *BookingRepository) GetByUserID(ctx context.Context, userID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("start_time DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) GetByID(ctx context.Context, bookingID uint) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.WithContext(ctx).First(&booking, bookingID).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) CancelBooking(ctx context.Context, bookingID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", bookingID).
		Update("status", constants.BookingStatusCancelled).Error
}
