package timeslot_test

import (
	"testing"
	"time"

	"restaurant-table-reservation/internal/timeslot"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValid(t *testing.T) {
	assert.True(t, timeslot.Valid("19:00"))
	assert.False(t, timeslot.Valid("15:30"))
}

func TestDateTime(t *testing.T) {
	date := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	dt, err := timeslot.DateTime(date, "19:00")
	require.NoError(t, err)
	assert.Equal(t, 19, dt.Hour())
	assert.Equal(t, 10, dt.Day())
}
