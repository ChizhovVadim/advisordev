package candleupdater

import (
	"advisordev/internal/core"
	"regexp"
	"strconv"
	"time"
)

var securityCodeRegexp = regexp.MustCompile("([A-Za-z]+)-([0-9]{1,2}).([0-9]{2})")

func parseSecurityCode(securityCode string) (security string, month, year int, ok bool) {
	res := securityCodeRegexp.FindStringSubmatch(securityCode)
	if res == nil {
		return
	}
	security = res[1]
	var err error
	month, err = strconv.Atoi(res[2])
	if err != nil {
		return
	}
	year, err = strconv.Atoi(res[3])
	if err != nil {
		return
	}
	var curYear = time.Now().Year()
	year += curYear - curYear%100
	ok = true
	return
}

func expirationDate(securityCode string) (time.Time, bool) {
	var _, m, y, ok = parseSecurityCode(securityCode)
	if !ok {
		return time.Time{}, false
	}
	var t = time.Date(y, time.Month(m), 15, 0, 0, 0, 0, time.Local)
	// TODO С 1 июля 2015, для новых серий по кот нет открытых позиций, все основные фьючерсы и опционы должны исполняться в 3-й четверг месяца
	return t, true
}

func skipEarly(source []core.Candle, time time.Time) []core.Candle {
	var index = -1
	for i := range source {
		if source[i].DateTime.After(time) {
			index = i
			break
		}
	}
	if index == -1 {
		return nil
	}
	return source[index:]
}
