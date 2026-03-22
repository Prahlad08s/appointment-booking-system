package repositories

import (
	"appointment-booking-system/config"
	"appointment-booking-system/models"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg *config.Config, logger *zap.Logger) *gorm.DB {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	logger.Info("Database connection established")

	if err := runMigrations(db, logger); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	DB = db
	return db
}

func runMigrations(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Running database migrations")

	if err := db.AutoMigrate(
		&models.Coach{},
		&models.CoachAvailability{},
		&models.Booking{},
	); err != nil {
		return err
	}

	// Partial unique index: only active bookings enforce uniqueness on (coach_id, start_time)
	// This is the DB-level safeguard against double booking
	result := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_booking 
		ON bookings(coach_id, start_time) WHERE status = 'booked'`)
	if result.Error != nil {
		return result.Error
	}

	logger.Info("Database migrations completed")
	return nil
}
