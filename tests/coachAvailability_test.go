package tests

import (
	"net/http"
	"testing"

	"appointment-booking-system/models"

	"github.com/stretchr/testify/assert"
)

func TestSetAvailability_Success(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Test Coach", "coach_avail@test.com", "UTC")

	body := models.SetAvailabilityRequest{
		CoachID:   coach.ID,
		DayOfWeek: 1, // Monday
		StartTime: "09:00",
		EndTime:   "14:00",
	}

	w := performRequest("POST", "/v1/coaches/availability", body)
	assert.Equal(t, http.StatusCreated, w.Code)

	resp := parseResponse(w)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "availability set successfully", resp.Message)
}

func TestSetAvailability_InvalidTimeRange(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Test Coach", "coach_avail_invalid@test.com", "UTC")

	body := models.SetAvailabilityRequest{
		CoachID:   coach.ID,
		DayOfWeek: 1,
		StartTime: "14:00",
		EndTime:   "09:00",
	}

	w := performRequest("POST", "/v1/coaches/availability", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	resp := parseResponse(w)
	assert.Equal(t, "error", resp.Status)
}

func TestSetAvailability_OverlappingWindows(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Test Coach", "coach_overlap@test.com", "UTC")
	seedAvailability(coach.ID, 1, "09:00:00", "12:00:00")

	body := models.SetAvailabilityRequest{
		CoachID:   coach.ID,
		DayOfWeek: 1,
		StartTime: "11:00",
		EndTime:   "15:00",
	}

	w := performRequest("POST", "/v1/coaches/availability", body)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGetAvailability_Success(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Test Coach", "coach_get_avail@test.com", "UTC")
	seedAvailability(coach.ID, 1, "09:00:00", "12:00:00")
	seedAvailability(coach.ID, 3, "13:00:00", "17:00:00")

	w := performRequest("GET", "/v1/coaches/availability?coach_id="+itoa(coach.ID), nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResponse(w)
	assert.Equal(t, "success", resp.Status)
}

func TestGetAvailability_MissingCoachID(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("GET", "/v1/coaches/availability", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
