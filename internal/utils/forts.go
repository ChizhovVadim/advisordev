package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TimeRange struct {
	StartYear     int
	StartQuarter  int
	FinishYear    int
	FinishQuarter int
}

func QuarterSecurityCodes(name string, tr TimeRange) []string {
	var result []string
	for year := tr.StartYear; year <= tr.FinishYear; year++ {
		for quarter := 0; quarter < 4; quarter++ {
			if year == tr.StartYear && quarter < tr.StartQuarter {
				continue
			}
			if year == tr.FinishYear && quarter > tr.FinishQuarter {
				break
			}
			var securityCode = fmt.Sprintf("%v-%v.%02d", name, 3+quarter*3, year%100)
			result = append(result, securityCode)
		}
	}
	return result
}

func ApproxExpirationDate(securityCode string) time.Time {
	// С 1 июля 2015, для новых серий по кот нет открытых позиций, все основные фьючерсы и опционы должны исполняться в 3-й четверг месяца
	// name-month.year
	var delim1 = strings.Index(securityCode, "-")
	if delim1 == -1 {
		return time.Time{}
	}
	var delim2 = strings.Index(securityCode, ".")
	if delim2 == -1 {
		return time.Time{}
	}
	month, err := strconv.Atoi(securityCode[delim1+1 : delim2])
	if err != nil {
		return time.Time{}
	}
	year, err := strconv.Atoi(securityCode[delim2+1:])
	if err != nil {
		return time.Time{}
	}
	var curYear = time.Now().Year()
	year = curYear - curYear%100 + year
	return time.Date(year, time.Month(month), 15, 0, 0, 0, 0, nil)
}

// Sample: "Si-3.17" -> "SiH7"
// http://moex.com/s205
func EncodeSecurity(securityName string) (string, error) {
	const MonthCodes = "FGHJKMNQUVXZ"
	var parts = strings.SplitN(securityName, "-", 2)
	var name = parts[0]
	parts = strings.SplitN(parts[1], ".", 2)
	month, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", err
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", err
	}

	//TODO hack
	if name == "CNY" {
		name = "CR"
	}

	return fmt.Sprintf("%v%v%v", name, string(MonthCodes[month-1]), year%10), nil
}
