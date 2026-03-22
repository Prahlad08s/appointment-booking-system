package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSlots_Success(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Slots Coach", "slots_coach@test.com", "UTC")

	// Find the next Monday from today
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "12:00:00")

	url := fmt.Sprintf("/v1/users/slots?coach_id=%d&date=%s", coach.ID, nextMonday.Format("2006-01-02"))
	w := performRequest("GET", url, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	resp := parseResponse(w)
	assert.Equal(t, "success", resp.Status)

	// 09:00 to 12:00 = 6 slots of 30 minutes
	slots, ok := resp.Data.([]interface{})
	if ok {
		assert.Equal(t, 6, len(slots))
	}
}

func TestGetSlots_NoAvailability(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Empty Coach", "empty_coach@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)

	url := fmt.Sprintf("/v1/users/slots?coach_id=%d&date=%s", coach.ID, nextMonday.Format("2006-01-02"))
	w := performRequest("GET", url, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSlots_InvalidDate(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("GET", "/v1/users/slots?coach_id=1&date=not-a-date", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSlots_PastDate(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("GET", "/v1/users/slots?coach_id=1&date=2020-01-01", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSlots_MissingCoachID(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("GET", "/v1/users/slots?date=2026-12-01", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSlots_MissingDate(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("GET", "/v1/users/slots?coach_id=1", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSlots_InvalidCoachID(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	w := performRequest("GET", "/v1/users/slots?coach_id=abc&date=2026-12-01", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSlots_WithTimezone(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("TZ Coach", "tz_coach@test.com", "America/New_York")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "11:00:00")

	url := fmt.Sprintf("/v1/users/slots?coach_id=%d&date=%s&timezone=America/Chicago",
		coach.ID, nextMonday.Format("2006-01-02"))
	w := performRequest("GET", url, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func findNextWeekday(day time.Weekday) time.Time {
	now := time.Now().UTC()
	daysUntil := int(day) - int(now.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return now.AddDate(0, 0, daysUntil)
}
