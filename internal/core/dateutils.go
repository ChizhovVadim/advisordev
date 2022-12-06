package core

import "time"

var Moscow = initMoscow()

func initMoscow() *time.Location {
	var loc, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.FixedZone("MSK", int(3*time.Hour/time.Second))
	}
	return loc
}

func DateTimeToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func FirstDayOMonth(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
}

func IsMainFortsSession(d time.Time) bool {
	return d.Hour() >= 10 && d.Hour() <= 18
}

func IsMainFortsSessionStarted(l, r time.Time) bool {
	return IsMainFortsSession(r) &&
		(!IsMainFortsSession(l) || IsNewDayStarted(l, r))
}

func IsNewDayStarted(l, r time.Time) bool {
	return (l.Year() != r.Year() || l.Month() != r.Month() || l.Day() != r.Day()) &&
		r.After(l)
}

// IsMainFortsSessionFinished
func IsNewFortsDateStarted(l, r time.Time) bool {
	return IsMainFortsSession(l) && (!IsMainFortsSession(r) || IsNewDayStarted(l, r))
}

func TimeOfDay(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute
}

func IsNewPeriod(l, r time.Time, period time.Duration) bool {
	return TimeOfDay(l) < period && period <= TimeOfDay(r)
}

func IsNewWeekStarted(l, r time.Time) bool {
	var py, pw = l.ISOWeek()
	var y, w = r.ISOWeek()
	return py != y || pw != w
}

func IsNewMonthStarted(l, r time.Time) bool {
	return l.Year() != r.Year() || l.Month() != r.Month()
}
