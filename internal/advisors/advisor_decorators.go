package advisors

import (
	"advisordev/internal/domain"
	"advisordev/internal/utils"
	"log/slog"
	"math"
)

func ApplyCandleValidation(advisor domain.Advisor, logger *slog.Logger) domain.Advisor {
	var lastCandle = domain.Candle{}
	var errorMode = false
	return func(candle domain.Candle) domain.Advice {
		if !lastCandle.DateTime.IsZero() {
			if !candle.DateTime.After(lastCandle.DateTime) {
				if !errorMode {
					errorMode = true
					logger.Warn("Invalid candle order",
						"prev", lastCandle,
						"cur", candle)
				}
				return domain.Advice{}
			}
			{
				var change = math.Log(candle.ClosePrice / lastCandle.ClosePrice)
				if math.Abs(change) >= 0.1 {
					logger.Warn("Big jump",
						"change", change,
						"prev", lastCandle,
						"cur", candle)
				}
			}
			/*if candle.DateTime.Hour() >= 19 {
				return domain.Advice{}
			}*/
			{
				if utils.IsNewDayStarted(lastCandle.DateTime, candle.DateTime) &&
					candle.DateTime.Minute() == 55 &&
					candle.DateTime.Hour() <= 9 {
					// аукцион открытия
					// полностью игнорируем, prevCandle не устанавливаем
					// для расчета волатильности учитывать нельзя, тк узкий диапазон
					// для торговли можно учитывать, но только если хороший объем (бывает вообще 1 контракт)
					return domain.Advice{}
				}
			}
		}
		errorMode = false //!
		lastCandle = candle
		return advisor(candle)
	}
}
