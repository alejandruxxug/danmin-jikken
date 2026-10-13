package game

import (
	"fmt"
	"strconv"
	"strings"
)

// Clock is the in-game time of day. The game moves it forward, so the
// agent backdoor can set it (3:17 AM on demand) and tests can read it.
type Clock struct {
	minutes float64 // minutes since midnight
}

// NewClock starts at hour:minute.
func NewClock(hour, minute int) *Clock {
	return &Clock{minutes: float64(hour*60 + minute)}
}

// ParseClock reads "HH:MM", for example "15:40".
func ParseClock(s string) (*Clock, error) {
	hh, mm, ok := strings.Cut(s, ":")
	if !ok {
		return nil, fmt.Errorf("clock %q: want HH:MM", s)
	}
	h, err := strconv.Atoi(hh)
	if err != nil || h < 0 || h > 23 {
		return nil, fmt.Errorf("clock %q: hour must be 0-23", s)
	}
	m, err := strconv.Atoi(mm)
	if err != nil || m < 0 || m > 59 {
		return nil, fmt.Errorf("clock %q: minute must be 0-59", s)
	}
	return NewClock(h, m), nil
}

// Advance moves the clock forward by the given number of in-game minutes.
func (c *Clock) Advance(minutes float64) {
	c.minutes += minutes
	for c.minutes >= 24*60 {
		c.minutes -= 24 * 60
	}
}

// String gives "HH:MM".
func (c *Clock) String() string {
	total := int(c.minutes)
	return fmt.Sprintf("%02d:%02d", total/60, total%60)
}
