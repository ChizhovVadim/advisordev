package history

import (
	"advisordev/internal/domain"
	"advisordev/internal/utils"
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
	MonthHpr           float64
	StDev              float64
	AVaR               float64
	DayHprs            []DateSum
	MonthHprs          []DateSum
	YearHprs           []DateSum
	DrawdownInfo       DrawdownInfo
	ProfitableRating   []DateSum
	UnprofitableRating []DateSum
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
	report.DrawdownInfo = computeDrawdownInfo(hprs)

	var sortedHprs = make([]DateSum, len(hprs))
	copy(sortedHprs, hprs)
	sort.Slice(sortedHprs, func(i, j int) bool {
		return sortedHprs[i].Sum < sortedHprs[j].Sum
	})
	report.AVaR = meanBySum(sortedHprs[:len(sortedHprs)/20])
	//TODO make copy
	report.ProfitableRating = sortedHprs[utils.Max(0, len(sortedHprs)-10):]
	report.UnprofitableRating = sortedHprs[:utils.Min(len(sortedHprs), 10)]

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
	return utils.StDev(x)
}

func meanBySum(hprs []DateSum) float64 {
	var items = make([]float64, len(hprs))
	for i := range items {
		items[i] = hprs[i].Sum
	}
	mean, _ := utils.Moments(items)
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
	PrintHprs(report.DayHprs[utils.Max(0, len(report.DayHprs)-20):])

	fmt.Println("Доходности по месяцам")
	PrintHprs(report.MonthHprs)

	fmt.Println("Доходности по годам")
	PrintHprs(report.YearHprs)

	fmt.Println("Самые прибыльные дни")
	PrintHprs(report.ProfitableRating)

	fmt.Println("Самые убыточные дни")
	PrintHprs(report.UnprofitableRating)
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

func PrintAdvices(advices []domain.Advice) {
	for _, item := range advices {
		fmt.Println(item)
	}
}
