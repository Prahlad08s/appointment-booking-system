package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"appointment-booking-system/commons/constants"
	"appointment-booking-system/models"

	"github.com/stretchr/testify/assert"
)

func TestCancelBooking_Success(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Cancel Coach", "cancel_coach@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "12:00:00")

	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		9, 0, 0, 0, time.UTC)
	body := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   coach.ID,
		StartTime: slotTime.Format(time.RFC3339),
	}

	w := performRequest("POST", "/v1/users/bookings", body)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Extract booking ID from response
	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	dataMap, _ := resp.Data.(map[string]interface{})
	bookingID := uint(dataMap["id"].(float64))

	// Cancel the booking
	cancelURL := fmt.Sprintf("/v1/users/bookings/%d?user_id=101", bookingID)
	w2 := performRequest("DELETE", cancelURL, nil)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestCancelBooking_SlotBecomesAvailable(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Rebook Coach", "rebook_coach@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "10:00:00") // 2 slots

	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		9, 0, 0, 0, time.UTC)
	bookBody := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   coach.ID,
		StartTime: slotTime.Format(time.RFC3339),
	}

	w := performRequest("POST", "/v1/users/bookings", bookBody)
	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	dataMap, _ := resp.Data.(map[string]interface{})
	bookingID := uint(dataMap["id"].(float64))

	// Verify only 1 slot remains
	slotsURL := fmt.Sprintf("/v1/users/slots?coach_id=%d&date=%s", coach.ID, nextMonday.Format("2006-01-02"))
	w2 := performRequest("GET", slotsURL, nil)
	resp2 := parseResponse(w2)
	slots, _ := resp2.Data.([]interface{})
	assert.Equal(t, 1, len(slots))

	// Cancel the booking
	cancelURL := fmt.Sprintf("/v1/users/bookings/%d?user_id=101", bookingID)
	performRequest("DELETE", cancelURL, nil)

	// Verify 2 slots are available again
	w3 := performRequest("GET", slotsURL, nil)
	resp3 := parseResponse(w3)
	slots3, _ := resp3.Data.([]interface{})
	assert.Equal(t, 2, len(slots3))
}

func TestCancelBooking_NotFound(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("DELETE", "/v1/users/bookings/99999?user_id=101", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCancelBooking_NotOwned(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Own Coach", "own_coach@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "12:00:00")

	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		9, 0, 0, 0, time.UTC)
	body := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   coach.ID,
		StartTime: slotTime.Format(time.RFC3339),
	}

	w := performRequest("POST", "/v1/users/bookings", body)
	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	dataMap, _ := resp.Data.(map[string]interface{})
	bookingID := uint(dataMap["id"].(float64))

	// Try to cancel with different user_id
	cancelURL := fmt.Sprintf("/v1/users/bookings/%d?user_id=999", bookingID)
	w2 := performRequest("DELETE", cancelURL, nil)
	assert.Equal(t, http.StatusForbidden, w2.Code)
}

func TestCancelBooking_AlreadyCancelled(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Dup Cancel Coach", "dup_cancel@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "12:00:00")

	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		9, 0, 0, 0, time.UTC)
	body := models.CreateBookingRequest{
		UserID:    101,
		CoachID:   coach.ID,
		StartTime: slotTime.Format(time.RFC3339),
	}

	w := performRequest("POST", "/v1/users/bookings", body)
	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	dataMap, _ := resp.Data.(map[string]interface{})
	bookingID := uint(dataMap["id"].(float64))

	cancelURL := fmt.Sprintf("/v1/users/bookings/%d?user_id=101", bookingID)
	performRequest("DELETE", cancelURL, nil)

	// Cancel again
	w2 := performRequest("DELETE", cancelURL, nil)
	assert.Equal(t, http.StatusConflict, w2.Code)

	resp2 := parseResponse(w2)
	assert.Equal(t, constants.ErrAlreadyCancelled, resp2.Message)
}

func TestCancelBooking_MissingUserID(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("DELETE", "/v1/users/bookings/1", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
