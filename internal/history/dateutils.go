package history

import (
	"time"
)

func DateTimeToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func IsNewDayStarted(l, r time.Time) bool {
	return (l.Year() != r.Year() || l.Month() != r.Month() || l.Day() != r.Day()) &&
		r.After(l)
}

func firstDayOfYear(d time.Time) time.Time {
	return time.Date(d.Year(), 1, 1, 0, 0, 0, 0, d.Location())
}

func firstDayOMonth(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
}

func IsAfterLongHolidays(l, r time.Time) bool {
	if !IsNewDayStarted(l, r) {
		return false
	}
	y, m, d := l.Date()
	if y == 2022 && m == time.February && d == 25 {
		// приостановка торгов из-за СВО. выйти заранее невозможно!
		return false
	}
	var startDate = DateTimeToDate(l).AddDate(0, 0, 1)
	var endDate = DateTimeToDate(r)
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		var weekDay = d.Weekday()
		if weekDay != time.Sunday && weekDay != time.Saturday {
			//В промежутке между currentDate и nextDate был 1 не выходной (значит торговала Америка)
			return true
		}
	}
	return false
}
