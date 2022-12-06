package history

import (
	"advisordev/internal/core"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

const dateFormatLayout = "2006-01-02"

type DateSum struct {
	Date time.Time
	Sum  float64
}

type HprStatistcs struct {
	MonthHpr     float64
	StDev        float64
	AVaR         float64
	DayHprs      []DateSum
	MonthHprs    []DateSum
	YearHprs     []DateSum
	DrawdownInfo DrawdownInfo
}

type DrawdownInfo struct {
	HighEquityDate      time.Time
	MaxDrawdown         float64
	LongestDrawdown     int
	CurrentDrawdown     float64
	CurrentDrawdownDays int
}

func ComputeHprStatistcs(hprs []DateSum) HprStatistcs {
	var report = HprStatistcs{}
	report.DayHprs = hprs
	report.MonthHprs = hprsByPeriod(hprs, firstDayOMonth)
	report.YearHprs = hprsByPeriod(hprs, firstDayOfYear)
	report.MonthHpr = math.Pow(totalHpr(hprs), 22.0/float64(len(hprs)))
	report.StDev = stDevHprs(hprs)
	report.AVaR = cvarHprs(hprs)
	report.DrawdownInfo = computeDrawdownInfo(hprs)
	return report
}

func computeDrawdownInfo(hprs []DateSum) DrawdownInfo {
	var currentSum = 0.0
	var maxSum = 0.0
	var longestDrawdown = 0
	var currentDrawdownDays = 0
	var maxDrawdown = 0.0
	var highEquityDate = hprs[0].Date

	for _, hpr := range hprs {
		currentSum += math.Log(hpr.Sum)
		if currentSum > maxSum {
			maxSum = currentSum
			highEquityDate = hpr.Date
		}
		if curDrawdownn := currentSum - maxSum; curDrawdownn < maxDrawdown {
			maxDrawdown = curDrawdownn
		}
		currentDrawdownDays = int(hpr.Date.Sub(highEquityDate) / (time.Hour * 24))
		if currentDrawdownDays > longestDrawdown {
			longestDrawdown = currentDrawdownDays
		}
	}

	return DrawdownInfo{
		HighEquityDate:      highEquityDate,
		LongestDrawdown:     longestDrawdown,
		CurrentDrawdownDays: currentDrawdownDays,
		MaxDrawdown:         math.Exp(maxDrawdown),
		CurrentDrawdown:     math.Exp(currentSum - maxSum),
	}
}

func totalHpr(source []DateSum) float64 {
	var result = 1.0
	for _, item := range source {
		result *= item.Sum
	}
	return result
}

func stDevHprs(source []DateSum) float64 {
	var x = make([]float64, len(source))
	for i := range source {
		x[i] = math.Log(source[i].Sum)
	}
	return core.StDev(x)
}

func cvarHprs(hprs []DateSum) float64 {
	var count = (len(hprs) - 1) / 20
	if count < 1 {
		return math.NaN()
	}
	var items = make([]float64, len(hprs))
	for i := range items {
		items[i] = hprs[i].Sum
	}
	sort.Float64s(items)
	mean, _ := core.Moments(items[:count])
	return mean
}

func hprsByPeriod(hprs []DateSum, period func(time.Time) time.Time) []DateSum {
	var result []DateSum
	for i, hpr := range hprs {
		if i == 0 || period(result[len(result)-1].Date) != period(hpr.Date) {
			result = append(result, hpr)
		} else {
			var item = &result[len(result)-1]
			item.Date = hpr.Date
			item.Sum *= hpr.Sum
		}
	}
	return result
}

func PrintHprReport(report HprStatistcs) {
	var w = newTabWriter()
	fmt.Fprintf(w, "Ежемесячная доходность\t%.1f%%\t\n", (report.MonthHpr-1)*100)
	fmt.Fprintf(w, "Среднеквадратичное отклонение доходности за день\t%.1f%%\t\n", report.StDev*100)
	fmt.Fprintf(w, "Средний убыток в день среди 5%% худших дней\t%.1f%%\t\n", (report.AVaR-1)*100)
	printDrawdownInfo(w, report.DrawdownInfo)
	w.Flush()

	fmt.Println("Доходности по дням")
	var dayHprs = report.DayHprs
	if skip := len(dayHprs) - 20; skip > 0 {
		dayHprs = dayHprs[skip:]
	}
	PrintHprs(dayHprs)

	fmt.Println("Доходности по месяцам")
	PrintHprs(report.MonthHprs)

	fmt.Println("Доходности по годам")
	PrintHprs(report.YearHprs)
}

func hprPercent(hpr float64) float64 {
	return (hpr - 1) * 100
}

func printDrawdownInfo(w io.Writer, info DrawdownInfo) {
	fmt.Fprintf(w, "Максимальная просадка\t%.1f%%\t\n", hprPercent(info.MaxDrawdown))
	fmt.Fprintf(w, "Продолжительная просадка\t%v дн.\t\n", info.LongestDrawdown)
	fmt.Fprintf(w, "Текущая просадка\t%.1f%% %v дн.\t\n", hprPercent(info.CurrentDrawdown), info.CurrentDrawdownDays)
	fmt.Fprintf(w, "Дата максимума эквити\t%v\t\n", info.HighEquityDate.Format(dateFormatLayout))
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
}

func PrintHprs(source []DateSum) {
	var w = newTabWriter()
	for _, item := range source {
		fmt.Fprintf(w, "%v\t%.1f%%\t\n", item.Date.Format(dateFormatLayout), (item.Sum-1)*100)
	}
	w.Flush()
}

func PrintAdvices(advices []core.Advice) {
	var w = newTabWriter()
	fmt.Fprintf(w, "Security\tTime\tPrice\tPosition\t\n")
	for _, item := range advices {
		fmt.Fprintf(w, "%v\t%v\t%v\t%.2f\t\n",
			item.SecurityCode, item.DateTime.Format("2006-01-02T15:04"), item.Price, item.Position)
	}
	w.Flush()
}
