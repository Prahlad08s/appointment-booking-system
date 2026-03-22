package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"github.com/stretchr/testify/assert"
)

func TestGetUserBookings_Success(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Bookings Coach", "bookings_coach@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "17:00:00")

	// Create two bookings for user 101
	for _, hour := range []int{9, 10} {
		slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
			hour, 0, 0, 0, time.UTC)
		body := models.CreateBookingRequest{
			UserID:    101,
			CoachID:   coach.ID,
			StartTime: slotTime.Format(time.RFC3339),
		}
		performRequest("POST", "/v1/users/bookings", body)
	}

	w := performRequest("GET", "/v1/users/bookings?user_id=101", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResponse(w)
	assert.Equal(t, "success", resp.Status)

	bookings, ok := resp.Data.([]interface{})
	if ok {
		assert.Equal(t, 2, len(bookings))
	}
}

func TestGetUserBookings_MissingUserID(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("GET", "/v1/users/bookings", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	resp := parseResponse(w)
	assert.Equal(t, constants.ErrMissingUserID, resp.Message)
}

func TestGetUserBookings_InvalidUserID(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("GET", "/v1/users/bookings?user_id=abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserBookings_EmptyResult(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("GET", "/v1/users/bookings?user_id=999", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResponse(w)
	assert.Equal(t, "success", resp.Status)
}

func itoa(n uint) string {
	return fmt.Sprintf("%d", n)
}
