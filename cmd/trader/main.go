package main

import (
	"advisordev/internal/candles"
	"advisordev/internal/cli"
	"advisordev/internal/moex"
	"advisordev/internal/trader"
	"flag"
	"fmt"
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
	var clientKey string

	flag.StringVar(&clientKey, "client", "", "client key")
	flag.Parse()

	settings, err := loadSettings(cli.MapPath("~/TradingData/luatrader.xml"))
	if err != nil {
		panic(err)
	}

	if clientKey == "" && len(settings.Clients) > 1 {
		fmt.Printf("Enter client: ")
		fmt.Scanln(&clientKey)
	}

	client, err := findClient(settings.Clients, clientKey)
	if err != nil {
		panic(err)
	}

	var today = time.Now()
	var logFilePath = buildLogFilePath(clientKey, today, "")

	// main log
	fLog, err := appendFile(logFilePath)
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
	)).With("client", clientKey)

	err = run(logger, settings, client)
	if err != nil {
		logger.Error("run failed",
			"error", err)
		return
	}
}

func run(
	logger *slog.Logger,
	settings Settings,
	client trader.Client,
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
	return trader.Run(logger, candleStorage, client, settings.StrategyConfigs)
}

func findClient(clients []trader.Client, clientKey string) (trader.Client, error) {
	if clientKey == "" && len(clients) == 1 {
		return clients[0], nil
	}
	for i := range clients {
		if clients[i].Key == clientKey {
			return clients[i], nil
		}
	}
	return trader.Client{}, fmt.Errorf("no client found %v", clientKey)
}

func buildLogFilePath(clientKey string, date time.Time, name string) string {
	var logFolderPath = filepath.Join(cli.MapPath("~/TradingData/Logs/luatrader"), clientKey)
	var err = os.MkdirAll(logFolderPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	var dateName = date.Format("2006-01-02")
	return filepath.Join(logFolderPath, dateName+name+".txt")
}

func appendFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}
