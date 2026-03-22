package models

type CreateCoachRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Timezone string `json:"timezone"`
}

type SetAvailabilityRequest struct {
	CoachID   uint   `json:"coach_id" binding:"required,gt=0"`
	DayOfWeek int    `json:"day_of_week" binding:"gte=0,lte=6"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

type CreateBookingRequest struct {
	UserID       uint   `json:"user_id" binding:"required,gt=0"`
	CoachID      uint   `json:"coach_id" binding:"required,gt=0"`
	StartTime    string `json:"start_time" binding:"required"`
	UserTimezone string `json:"user_timezone"`
}
