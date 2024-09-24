package utils

import "time"

var Moscow = initMoscow()

func initMoscow() *time.Location {
	var loc, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.FixedZone("MSK", int(3*time.Hour/time.Second))
	}
	return loc
}

func FromOneDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func IsNewFortsDateStarted(l, r time.Time) bool {
	return IsMainFortsSession(l) && (!IsMainFortsSession(r) || IsNewDayStarted(l, r))
}

func IsNewDayStarted(l, r time.Time) bool {
	return (l.Year() != r.Year() || l.Month() != r.Month() || l.Day() != r.Day()) &&
		r.After(l)
}

func IsMainFortsSession(d time.Time) bool {
	return d.Hour() >= 10 && d.Hour() <= 18
}

func TimeOfDay(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute
}

// TODO разные даты l, r
func IsNewPeriod(l, r time.Time, period time.Duration) bool {
	return TimeOfDay(l) < period && period <= TimeOfDay(r)
}
