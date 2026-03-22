package commons

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSlots_ThreeHourWindow(t *testing.T) {
	date := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) // A Monday
	start, _ := time.Parse("15:04", "09:00")
	end, _ := time.Parse("15:04", "12:00")

	slots := GenerateSlots(date, start, end, time.UTC)

	// 09:00-12:00 = 6 slots of 30 min
	assert.Equal(t, 6, len(slots))
	assert.Equal(t, time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC), slots[0].StartTime)
	assert.Equal(t, time.Date(2026, 10, 26, 9, 30, 0, 0, time.UTC), slots[0].EndTime)
	assert.Equal(t, time.Date(2026, 10, 26, 11, 30, 0, 0, time.UTC), slots[5].StartTime)
	assert.Equal(t, time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC), slots[5].EndTime)
}

func TestGenerateSlots_OneHourWindow(t *testing.T) {
	date := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC)
	start, _ := time.Parse("15:04", "14:00")
	end, _ := time.Parse("15:04", "15:00")

	slots := GenerateSlots(date, start, end, time.UTC)

	assert.Equal(t, 2, len(slots))
}

func TestGenerateSlots_LessThan30Min(t *testing.T) {
	date := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC)
	start, _ := time.Parse("15:04", "14:00")
	end, _ := time.Parse("15:04", "14:20")

	slots := GenerateSlots(date, start, end, time.UTC)

	assert.Equal(t, 0, len(slots))
}

func TestGenerateSlots_TimezoneConversion(t *testing.T) {
	date := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC)
	start, _ := time.Parse("15:04", "09:00")
	end, _ := time.Parse("15:04", "10:00")
	nyLoc, _ := time.LoadLocation("America/New_York")

	slots := GenerateSlots(date, start, end, nyLoc)

	// Coach is in New York (UTC-4 in October), so 09:00 EDT = 13:00 UTC
	assert.Equal(t, 2, len(slots))
	assert.Equal(t, 13, slots[0].StartTime.Hour()) // UTC
}

func TestFilterBookedSlots(t *testing.T) {
	slots := []TimeSlot{
		{StartTime: time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC), EndTime: time.Date(2026, 10, 26, 9, 30, 0, 0, time.UTC)},
		{StartTime: time.Date(2026, 10, 26, 9, 30, 0, 0, time.UTC), EndTime: time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)},
		{StartTime: time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC), EndTime: time.Date(2026, 10, 26, 10, 30, 0, 0, time.UTC)},
	}

	booked := map[time.Time]bool{
		time.Date(2026, 10, 26, 9, 30, 0, 0, time.UTC): true,
	}

	available := FilterBookedSlots(slots, booked)
	assert.Equal(t, 2, len(available))
	assert.Equal(t, 9, available[0].StartTime.Hour())
	assert.Equal(t, 10, available[1].StartTime.Hour())
}

func TestIsSlotAligned(t *testing.T) {
	assert.True(t, IsSlotAligned(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)))
	assert.True(t, IsSlotAligned(time.Date(2026, 1, 1, 9, 30, 0, 0, time.UTC)))
	assert.False(t, IsSlotAligned(time.Date(2026, 1, 1, 9, 15, 0, 0, time.UTC)))
	assert.False(t, IsSlotAligned(time.Date(2026, 1, 1, 9, 0, 30, 0, time.UTC)))
}

func TestParseTimeString(t *testing.T) {
	// HH:MM format
	t1, err := ParseTimeString("09:00")
	assert.NoError(t, err)
	assert.Equal(t, 9, t1.Hour())
	assert.Equal(t, 0, t1.Minute())

	// HH:MM:SS format
	t2, err := ParseTimeString("14:30:00")
	assert.NoError(t, err)
	assert.Equal(t, 14, t2.Hour())
	assert.Equal(t, 30, t2.Minute())

	// Invalid
	_, err = ParseTimeString("invalid")
	assert.Error(t, err)
}

func TestNormalizeTimeString(t *testing.T) {
	result, err := NormalizeTimeString("09:00")
	assert.NoError(t, err)
	assert.Equal(t, "09:00:00", result)

	result2, err := NormalizeTimeString("14:30:00")
	assert.NoError(t, err)
	assert.Equal(t, "14:30:00", result2)
}

func TestParseTimezone(t *testing.T) {
	loc, err := ParseTimezone("")
	assert.NoError(t, err)
	assert.Equal(t, time.UTC, loc)

	loc2, err := ParseTimezone("America/New_York")
	assert.NoError(t, err)
	assert.NotNil(t, loc2)

	_, err = ParseTimezone("Invalid/Zone")
	assert.Error(t, err)
}

func TestConvertSlotsToTimezone(t *testing.T) {
	nyLoc, _ := time.LoadLocation("America/New_York")
	slots := []TimeSlot{
		{
			StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 6, 15, 14, 30, 0, 0, time.UTC),
		},
	}

	converted := ConvertSlotsToTimezone(slots, nyLoc)
	assert.Equal(t, 10, converted[0].StartTime.Hour()) // 14:00 UTC = 10:00 EDT
}
