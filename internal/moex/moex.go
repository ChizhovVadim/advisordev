package moex

import (
	"advisordev/internal/domain"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const FuturesClassCode = "SPBFUT"

var TimeZone = initMoscow()

func initMoscow() *time.Location {
	var loc, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.FixedZone("MSK", int(3*time.Hour/time.Second))
	}
	return loc
}

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
	return time.Date(year, time.Month(month), 15, 0, 0, 0, 0, time.Local)
}

// Sample: "Si-3.17" -> "SiH7"
// http://moex.com/s205
func EncodeSecurity(securityName string) (string, error) {
	// вечные фьючерсы
	if strings.HasSuffix(securityName, "F") {
		return securityName, nil
	}

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

	// курс китайский юань – российский рубль
	if name == "CNY" {
		name = "CR"
	}

	return fmt.Sprintf("%v%v%v", name, string(MonthCodes[month-1]), year%10), nil
}

type SecurityInformator struct{}

func NewFortsSecurityInformator() *SecurityInformator {
	return &SecurityInformator{}
}

func (si *SecurityInformator) GetSecurityInfo(securityName string) (domain.SecurityInfo, error) {
	// квартальные фьючерсы
	if strings.HasPrefix(securityName, "Si") {
		securityCode, err := EncodeSecurity(securityName)
		if err != nil {
			return domain.SecurityInfo{}, err
		}
		return domain.SecurityInfo{
			Name:           securityName,
			ClassCode:      FuturesClassCode,
			Code:           securityCode,
			PricePrecision: 0,
			PriceStep:      1,
			PriceStepCost:  1,
			Lever:          1,
		}, nil
	}
	if strings.HasPrefix(securityName, "CNY") {
		securityCode, err := EncodeSecurity(securityName) //CR
		if err != nil {
			return domain.SecurityInfo{}, err
		}
		return domain.SecurityInfo{
			Name:           securityName,
			ClassCode:      FuturesClassCode,
			Code:           securityCode,
			PricePrecision: 3,
			PriceStep:      0.001,
			PriceStepCost:  1,
			Lever:          1000,
		}, nil
	}
	return domain.SecurityInfo{}, fmt.Errorf("secInfo not found %v", securityName)
}

func IsMainFortsSession(d time.Time) bool {
	return d.Hour() >= 10 && d.Hour() <= 18
}
