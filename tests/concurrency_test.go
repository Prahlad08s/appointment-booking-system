package tests

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"appointment-booking-system/models"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentBooking spawns multiple goroutines all trying to book the exact
// same slot at the same time. Only one should succeed (201), the rest should
// receive a 409 Conflict.
func TestConcurrentBooking_OnlyOneSucceeds(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Concurrent Coach", "concurrent@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "17:00:00")

	slotTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		10, 0, 0, 0, time.UTC)

	concurrentUsers := 10
	var wg sync.WaitGroup
	results := make(chan int, concurrentUsers)

	wg.Add(concurrentUsers)
	for i := 0; i < concurrentUsers; i++ {
		go func(userID int) {
			defer wg.Done()
			body := models.CreateBookingRequest{
				UserID:    uint(userID + 200),
				CoachID:   coach.ID,
				StartTime: slotTime.Format(time.RFC3339),
			}
			w := performRequest("POST", "/v1/users/bookings", body)
			results <- w.Code
		}(i)
	}

	wg.Wait()
	close(results)

	successCount := 0
	conflictCount := 0
	for code := range results {
		switch code {
		case http.StatusCreated:
			successCount++
		case http.StatusConflict:
			conflictCount++
		}
	}

	// Exactly 1 booking should succeed
	assert.Equal(t, 1, successCount, "Exactly one booking should succeed")
	// The rest should be conflicts
	assert.Equal(t, concurrentUsers-1, conflictCount, "All other bookings should be conflicts")
}

// TestConcurrentBooking_DifferentSlots ensures parallel bookings for different
// slots all succeed without interference.
func TestConcurrentBooking_DifferentSlots(t *testing.T) {
	setupTestRouter()
	defer cleanupDB()

	coach := seedCoach("Multi Slot Coach", "multi_slot@test.com", "UTC")
	nextMonday := findNextWeekday(time.Monday)
	seedAvailability(coach.ID, int(time.Monday), "09:00:00", "17:00:00")

	slots := []time.Time{
		time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 9, 0, 0, 0, time.UTC),
		time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 9, 30, 0, 0, time.UTC),
		time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 10, 0, 0, 0, time.UTC),
		time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 10, 30, 0, 0, time.UTC),
		time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 11, 0, 0, 0, time.UTC),
	}

	var wg sync.WaitGroup
	results := make(chan int, len(slots))

	wg.Add(len(slots))
	for i, slot := range slots {
		go func(userID int, slotTime time.Time) {
			defer wg.Done()
			body := models.CreateBookingRequest{
				UserID:    uint(userID + 300),
				CoachID:   coach.ID,
				StartTime: slotTime.Format(time.RFC3339),
			}
			w := performRequest("POST", "/v1/users/bookings", body)
			results <- w.Code
		}(i, slot)
	}

	wg.Wait()
	close(results)

	successCount := 0
	for code := range results {
		if code == http.StatusCreated {
			successCount++
		}
	}

	assert.Equal(t, len(slots), successCount, "All bookings for different slots should succeed")
}
