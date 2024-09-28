package main

import (
	"advisordev/internal/trader"
	"advisordev/internal/utils"
	"context"
	"flag"
	"fmt"
	"io"
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

type LogWrapper struct {
	logger *slog.Logger
	closer io.Closer
}

func main() {
	var logWrapper = &LogWrapper{
		logger: slog.New(slog.NewJSONHandler(os.Stderr, nil)),
	}
	defer func() {
		if logWrapper.closer != nil {
			logWrapper.closer.Close()
		}
	}()
	var err = run(logWrapper)
	if err != nil {
		logWrapper.logger.Error("app failed",
			"error", err)
	}
}

func run(logWrapper *LogWrapper) error {
	var clientKey string
	var quietMode bool

	flag.StringVar(&clientKey, "client", "", "client key")
	flag.BoolVar(&quietMode, "quiet", quietMode, "")
	flag.Parse()

	settings, err := loadSettings(utils.MapPath("./luatrader.xml"))
	if err != nil {
		return err
	}

	if clientKey == "" && len(settings.Clients) > 1 {
		fmt.Printf("Enter client: ")
		fmt.Scanln(&clientKey)
	}

	client, err := findClient(settings.Clients, clientKey)
	if err != nil {
		return err
	}

	var today = time.Now()

	err = initLogger(logWrapper, client.Key, today)
	if err != nil {
		return err
	}

	var logger = logWrapper.logger

	logger.Info("Application started.")
	defer logger.Info("Application closed.")

	logger.Info("Environment",
		"BuildDate", buildDate,
		"GitRevision", gitRevision)

	logger.Info("runtime",
		"Version", runtime.Version(),
		"NumCPU", runtime.NumCPU(),
		"GOMAXPROCS", runtime.GOMAXPROCS(0))

	// quik message log
	/*fQuikLog, err := os.OpenFile(buildLogFilePath(client.Key, today, "quik"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fQuikLog.Close()
	var quikLogger = log.New(fQuikLog, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// quik callback message log
	fQuikCallback, err := os.OpenFile(buildLogFilePath(client.Key, today, "quikcallback"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fQuikCallback.Close()
	var quikCallbackLogger = log.New(fQuikCallback, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)*/

	return trader.Run(context.Background(), logger, settings.StrategyConfigs, client, quietMode)
}

func buildLogFilePath(clientKey string, date time.Time, name string) string {
	var logFolderPath = filepath.Join(utils.MapPath("~/TradingData/Logs/luatrader"), clientKey)
	var dateName = date.Format("2006-01-02")
	return filepath.Join(logFolderPath, dateName+name+".log")
}

func initLogger(logWrapper *LogWrapper, clientKey string, date time.Time) error {
	var logFilePath = buildLogFilePath(clientKey, date, "")

	/*err := os.MkdirAll(logFolderPath, os.ModePerm)
	if err != nil {
		return err
	}*/

	// main log
	fLog, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	logWrapper.closer = fLog //вместо defer fLog.Close()

	// подменяем logger
	logWrapper.logger = slog.New(Fanout(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
		slog.NewJSONHandler(fLog, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)).With("client", clientKey)

	return nil
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
