package models

import (
	"time"
)

type Coach struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Timezone  string    `gorm:"type:varchar(100);default:'UTC'" json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CoachAvailability struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CoachID   uint      `gorm:"not null;index:idx_availability_coach" json:"coach_id"`
	Coach     Coach     `gorm:"foreignKey:CoachID;constraint:OnDelete:CASCADE" json:"-"`
	DayOfWeek int       `gorm:"not null;index:idx_availability_coach" json:"day_of_week"` // 0=Sunday..6=Saturday
	StartTime string    `gorm:"type:time;not null" json:"start_time"`
	EndTime   string    `gorm:"type:time;not null" json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Booking struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index:idx_bookings_user" json:"user_id"`
	CoachID      uint      `gorm:"not null" json:"coach_id"`
	Coach        Coach     `gorm:"foreignKey:CoachID;constraint:OnDelete:CASCADE" json:"-"`
	StartTime    time.Time `gorm:"not null" json:"start_time"`
	EndTime      time.Time `gorm:"not null" json:"end_time"`
	Status       string    `gorm:"type:varchar(20);not null;default:'booked';index:idx_bookings_user" json:"status"`
	UserTimezone string    `gorm:"type:varchar(100);default:'UTC'" json:"user_timezone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
