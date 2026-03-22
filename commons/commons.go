package commons

import (
	"fmt"
	"time"

	"appointment-booking-system/commons/constants"
)

type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// GenerateSlots produces 30-minute slots between start and end times on a given date.
// availStart and availEnd are clock times (e.g. 09:00, 14:00) in the coach's timezone.
// date is the calendar date. coachLocation is the coach's timezone.
// All returned slots are in UTC.
func GenerateSlots(date time.Time, availStart, availEnd time.Time, coachLocation *time.Location) []TimeSlot {
	slotDuration := time.Duration(constants.SlotDurationMinutes) * time.Minute

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(),
		availStart.Hour(), availStart.Minute(), 0, 0, coachLocation)
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(),
		availEnd.Hour(), availEnd.Minute(), 0, 0, coachLocation)

	var slots []TimeSlot
	for current := startOfDay; current.Add(slotDuration).Before(endOfDay) || current.Add(slotDuration).Equal(endOfDay); current = current.Add(slotDuration) {
		slots = append(slots, TimeSlot{
			StartTime: current.UTC(),
			EndTime:   current.Add(slotDuration).UTC(),
		})
	}

	return slots
}

// FilterBookedSlots removes slots that overlap with already booked times.
func FilterBookedSlots(allSlots []TimeSlot, bookedStartTimes map[time.Time]bool) []TimeSlot {
	var available []TimeSlot
	for _, slot := range allSlots {
		if !bookedStartTimes[slot.StartTime] {
			available = append(available, slot)
		}
	}
	return available
}

// ConvertSlotsToTimezone converts UTC slots to the requested timezone for display.
func ConvertSlotsToTimezone(slots []TimeSlot, loc *time.Location) []TimeSlot {
	converted := make([]TimeSlot, len(slots))
	for i, slot := range slots {
		converted[i] = TimeSlot{
			StartTime: slot.StartTime.In(loc),
			EndTime:   slot.EndTime.In(loc),
		}
	}
	return converted
}

// IsSlotAligned checks if a time is aligned to 30-minute boundaries (:00 or :30).
func IsSlotAligned(t time.Time) bool {
	return t.Minute()%constants.SlotDurationMinutes == 0 && t.Second() == 0
}

// ParseTimezone loads a timezone location, defaulting to UTC if empty.
func ParseTimezone(tz string) (*time.Location, error) {
	if tz == "" {
		return time.UTC, nil
	}
	return time.LoadLocation(tz)
}

// ParseTimeString accepts both "HH:MM" and "HH:MM:SS" formats and returns a parsed time.
func ParseTimeString(timeStr string) (time.Time, error) {
	t, err := time.Parse(constants.TimeFormat, timeStr)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("15:04", timeStr)
	if err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("time must be in HH:MM or HH:MM:SS format")
}

// NormalizeTimeString ensures time string is always stored as HH:MM:SS.
func NormalizeTimeString(timeStr string) (string, error) {
	t, err := ParseTimeString(timeStr)
	if err != nil {
		return "", err
	}
	return t.Format(constants.TimeFormat), nil
}
