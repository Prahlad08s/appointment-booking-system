package models

import "time"

type SlotResponse struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type BookingResponse struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	CoachID      uint      `json:"coach_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `json:"status"`
	UserTimezone string    `json:"user_timezone"`
	CreatedAt    time.Time `json:"created_at"`
}

type AvailabilityResponse struct {
	ID        uint   `json:"id"`
	CoachID   uint   `json:"coach_id"`
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
