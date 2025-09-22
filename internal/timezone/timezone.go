package timezone

import (
	"fmt"
	"time"
)

type Manager struct {
	location *time.Location
	name     string
}

func NewManager(tzName string) (*Manager, error) {
	if tzName == "" {
		tzName = "UTC"
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, err
	}

	return &Manager{
		loc, tzName,
	}, nil
}

// GetUTCNow returns current time in UTC
func (m *Manager) GetUTCNow() time.Time {
	return time.Now().UTC()
}

// ConvertToLocal converts UTC time to the configured local timezone
func (m *Manager) ConvertToLocal(utcTime time.Time) time.Time {
	return utcTime.In(m.location)
}

// ConvertToUTC converts a local time to UTC
func (m *Manager) ConvertToUTC(localTime time.Time) time.Time {
	if localTime.Location() != time.UTC {
		return localTime.UTC()
	}

	localTimeInTZ := time.Date(
		localTime.Year(), localTime.Month(), localTime.Day(),
		localTime.Hour(), localTime.Minute(), localTime.Second(),
		localTime.Nanosecond(), m.location,
	)

	return localTimeInTZ.UTC()
}

// ParseDatetimeLocal parses datetime-local input and converts to UTC
func (m *Manager) ParseDatetimeLocal(datetimeStr string) (time.Time, error) {
	localTime, err := time.Parse("2006-01-02T15:04", datetimeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid datetime format: %w", err)
	}

	localTimeInTZ := time.Date(
		localTime.Year(), localTime.Month(), localTime.Day(),
		localTime.Hour(), localTime.Minute(), localTime.Second(),
		0, m.location,
	)

	return localTimeInTZ.UTC(), nil
}

// FormatSmart provides smart formatting (time for today, date for others)
func (m *Manager) FormatSmart(utcTime time.Time) string {
	localTime := utcTime.In(m.location)
	now := time.Now().In(m.location)

	if localTime.Format("2006-01-02") == now.Format("2006-01-02") {
		return localTime.Format("15:04")
	}

	return localTime.Format("02.01.06")
}

// FormatFull provides full date and time formatting
func (m *Manager) FormatFull(utcTime time.Time) string {
	localTime := utcTime.In(m.location)
	return localTime.Format("02.01.06 15:04")
}

// FormatDateOnly provides date-only formatting
func (m *Manager) FormatDateOnly(utcTime time.Time) string {
	localTime := utcTime.In(m.location)
	return localTime.Format("02.01.06")
}

// FormatForDatetimeLocal formats time for datetime-local input
func (m *Manager) FormatForDatetimeLocal(utcTime time.Time) string {
	localTime := utcTime.In(m.location)
	return localTime.Format("2006-01-02T15:04")
}

// FormatLong provides a long, readable format
func (m *Manager) FormatLong(utcTime time.Time) string {
	localTime := utcTime.In(m.location)
	return localTime.Format("January 2, 2006 at 3:04 PM")
}

// GetTimezoneName returns the configured timezone name
func (m *Manager) GetTimezoneName() string {
	return m.name
}

// GetTimezoneOffset returns the current offset from UTC in minutes
func (m *Manager) GetTimezoneOffset() int {
	_, offset := time.Now().In(m.location).Zone()
	return offset / 60
}

// IsToday checks if the given UTC time is today in the local timezone
func (m *Manager) IsToday(utcTime time.Time) bool {
	localTime := utcTime.In(m.location)
	now := time.Now().In(m.location)
	return localTime.Format("2006-01-02") == now.Format("2006-01-02")
}

var TZ *Manager

func Init(timezoneName string) error {
	tz, err := NewManager(timezoneName)
	if err != nil {
		return err
	}
	TZ = tz
	return nil
}
