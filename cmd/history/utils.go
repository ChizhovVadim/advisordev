package main

import (
	"advisordev/internal/core"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func QuarterSecurityCodes(name string, startYear, startQuarter, finishYear, finishQuarter int) []string {
	var result []string
	for year := startYear; year <= finishYear; year++ {
		for quarter := 0; quarter < 4; quarter++ {
			if year == startYear && quarter < startQuarter {
				continue
			}
			if year == finishYear && quarter > finishQuarter {
				break
			}
			var securityCode = fmt.Sprintf("%v-%v.%02d", name, 3+quarter*3, year%100)
			result = append(result, securityCode)
		}
	}
	return result
}

func IsAfterLongHolidays(l, r time.Time) bool {
	if !core.IsNewDayStarted(l, r) {
		return false
	}
	y, m, d := l.Date()
	if y == 2022 && m == time.February && d == 25 {
		// приостановка торгов из-за СВО. выйти заранее невозможно!
		return false
	}
	var startDate = core.DateTimeToDate(l).AddDate(0, 0, 1)
	var endDate = core.DateTimeToDate(r)
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		var weekDay = d.Weekday()
		if weekDay != time.Sunday && weekDay != time.Saturday {
			//В промежутке между currentDate и nextDate был 1 не выходной (значит торговала Америка)
			return true
		}
	}
	return false
}

func MapPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		curUser, err := user.Current()
		if err != nil {
			return path
		}
		return filepath.Join(curUser.HomeDir, strings.TrimPrefix(path, "~/"))
	}
	if strings.HasPrefix(path, "./") {
		var exePath, err = os.Executable()
		if err != nil {
			return path
		}
		return filepath.Join(filepath.Dir(exePath), strings.TrimPrefix(path, "./"))
	}
	return path
}
