package timeslot

import (
	"fmt"
	"time"
)

// Fixed restaurant time slots (assumption documented in README/command.txt).
var All = []string{
	"12:00", "13:00", "14:00",
	"18:00", "19:00", "20:00", "21:00",
}

func Valid(slot string) bool {
	for _, s := range All {
		if s == slot {
			return true
		}
	}
	return false
}

func DateTime(date time.Time, slot string) (time.Time, error) {
	if !Valid(slot) {
		return time.Time{}, fmt.Errorf("invalid time slot: %s", slot)
	}

	parsed, err := time.Parse("15:04", slot)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(
		date.Year(), date.Month(), date.Day(),
		parsed.Hour(), parsed.Minute(), 0, 0,
		date.Location(),
	), nil
}
