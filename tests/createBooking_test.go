package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"appointment-booking-system/models"

	"github.com/stretchr/testify/assert"
)

func TestCreateBooking_Success(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Booking Coach", "booking_coach@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "12:00:00")

	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		9, 30, 0, 0, time.UTC)

	body := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   coach.ID,
		StartTime: slotTime.Format(time.RFC3339),
	}

	w := performRequest("POST", "/v1/users/bookings", body)
	assert.Equal(t, http.StatusCreated, w.Code)

	resp := parseResponse(w)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "booking created successfully", resp.Message)
}

func TestCreateBooking_SlotNotAligned(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Misaligned Coach", "misaligned@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "12:00:00")

	// 09:15 is not aligned to 30-min boundary
	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		9, 15, 0, 0, time.UTC)

	body := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   coach.ID,
		StartTime: slotTime.Format(time.RFC3339),
	}

	w := performRequest("POST", "/v1/users/bookings", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBooking_OutsideAvailability(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Outside Coach", "outside@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "12:00:00")

	// 14:00 is outside 09:00-12:00
	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		14, 0, 0, 0, time.UTC)

	body := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   coach.ID,
		StartTime: slotTime.Format(time.RFC3339),
	}

	w := performRequest("POST", "/v1/users/bookings", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBooking_DoubleBookingPrevented(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Double Coach", "double@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "12:00:00")

	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		10, 0, 0, 0, time.UTC)

	body := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   coach.ID,
		StartTime: slotTime.Format(time.RFC3339),
	}

	// First booking should succeed
	w1 := performRequest("POST", "/v1/users/bookings", body)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Second booking for same slot should fail with 409 Conflict
	body.UserID = 102
	w2 := performRequest("POST", "/v1/users/bookings", body)
	assert.Equal(t, http.StatusConflict, w2.Code)

	resp := parseResponse(w2)
	assert.Equal(t, "error", resp.Status)
}

func TestCreateBooking_SlotRemovedFromAvailable(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Removal Coach", "removal@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "10:00:00") // 2 slots

	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		9, 0, 0, 0, time.UTC)

	body := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   coach.ID,
		StartTime: slotTime.Format(time.RFC3339),
	}

	performRequest("POST", "/v1/users/bookings", body)

	// Check available slots - should be 1 (only 09:30 remaining)
	url := fmt.Sprintf("/v1/users/slots?coach_id=%d&date=%s", coach.ID, nextMonday.Format("2006-01-02"))
	w := performRequest("GET", url, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResponse(w)
	slots, ok := resp.Data.([]interface{})
	if ok {
		assert.Equal(t, 1, len(slots))
	}
}

func TestCreateBooking_PastSlot(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	body := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   1,
		StartTime: "2020-01-01T09:00:00Z",
	}

	w := performRequest("POST", "/v1/users/bookings", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
