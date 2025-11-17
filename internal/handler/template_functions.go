package handler

import (
	"fmt"
	"html/template"
	"time"

	"github.com/NiClassic/go-cloud/internal/timezone"
)

// GetTemplateFunctions returns custom template functions for timezone handling
func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"utcNow": func() time.Time {
			return timezone.TZ.GetUTCNow()
		},
		"addHours": func(t time.Time, hours int) time.Time {
			return t.Add(time.Duration(hours) * time.Hour)
		},
		"addDays": func(t time.Time, days int) time.Time {
			return t.AddDate(0, 0, days)
		},

		"formatSmart": func(t time.Time) string {
			return timezone.TZ.FormatSmart(t)
		},
		"formatFull": func(t time.Time) string {
			return timezone.TZ.FormatFull(t)
		},
		"formatDateOnly": func(t time.Time) string {
			return timezone.TZ.FormatDateOnly(t)
		},
		"formatLong": func(t time.Time) string {
			return timezone.TZ.FormatLong(t)
		},
		"formatDatetimeLocal": func(t time.Time) string {
			return timezone.TZ.FormatForDatetimeLocal(t)
		},
		"humanReadableSize": func(b int64) string {
			if b == -1 {
				return "-"
			}
			const unit = 1024
			if b < unit {
				return fmt.Sprintf("%d B", b)
			}
			div, exp := int64(unit), 0
			for n := b / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
		},
		"toLocalTime": func(t time.Time) time.Time {
			return timezone.TZ.ConvertToLocal(t)
		},

		"isToday": func(t time.Time) bool {
			return timezone.TZ.IsToday(t)
		},
		"timezoneName": func() string {
			return timezone.TZ.GetTimezoneName()
		},
		"timezoneOffset": func() int {
			return timezone.TZ.GetTimezoneOffset()
		},

		"isSameDay": func(t1, t2 time.Time) bool {
			local1 := timezone.TZ.ConvertToLocal(t1)
			local2 := timezone.TZ.ConvertToLocal(t2)
			return local1.Format("2006-01-02") == local2.Format("2006-01-02")
		},
		"daysDiff": func(t1, t2 time.Time) int {
			local1 := timezone.TZ.ConvertToLocal(t1)
			local2 := timezone.TZ.ConvertToLocal(t2)
			return int(local1.Sub(local2).Hours() / 24)
		},
	}
}
