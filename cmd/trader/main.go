package main

import (
	"advisordev/internal/candles"
	"advisordev/internal/cli"
	"advisordev/internal/moex"
	"advisordev/internal/trader"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	gitRevision string
	buildDate   string
)

func main() {
	settings, err := trader.LoadConfig("trader.xml")
	if err != nil {
		panic(err)
	}

	fLog, err := appendFile(buildLogFilePath(time.Now()))
	if err != nil {
		panic(err)
	}
	defer fLog.Close()

	var logger = slog.New(cli.Fanout(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
		slog.NewJSONHandler(fLog, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: false,
		}),
	))

	//slog.SetDefault(logger)

	err = run(logger, settings)
	if err != nil {
		logger.Error("run failed",
			"error", err)
		return
	}
}

func run(
	logger *slog.Logger,
	config trader.TraderConfig,
) error {
	logger.Debug("Application started.")
	defer logger.Debug("Application closed.")

	logger.Debug("Environment",
		"BuildDate", buildDate,
		"GitRevision", gitRevision)

	logger.Debug("runtime",
		"Version", runtime.Version(),
		"NumCPU", runtime.NumCPU(),
		"GOMAXPROCS", runtime.GOMAXPROCS(0))

	// временно такой путь:
	var candleStorage = candles.NewCandleStorageByPath(cli.MapPath("~/TradingData/Forts"), moex.TimeZone)
	//var candleInterval = domain.CandleIntervalMinutes5
	//var candleStorage = candles.NewCandleStorage(cli.MapPath("~/TradingData"), candleInterval, moex.TimeZone)
	return trader.Run(logger, candleStorage, config)
}

func buildLogFilePath(date time.Time) string {
	var logFolderPath = cli.MapPath("~/TradingData/Logs/luatrader")
	var err = os.MkdirAll(logFolderPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	var dateName = date.Format("2006-01-02")
	return filepath.Join(logFolderPath, dateName+".txt")
}

func appendFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}
