package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/factory"
	"github.com/urfave/cli"
)

var log = logger.GetOrCreate("mx-chain-sovereign-notifier")

const (
	configPath = "config/config.toml"

	logsPath       = "logs"
	logFilePrefix  = "sovereign-notifier"
	logLifeSpanSec = 432000 // 5 days
	logLifeSpanMb  = 1024   // 1 GB
)

func main() {
	app := cli.NewApp()
	app.Name = "MultiversX sovereign chain notifier"
	app.Usage = "This tool will communicate with an observer/light client connected to mx-chain via " +
		"websocket outport driver and listen to incoming transaction to the specified sovereign chain. If such transactions" +
		"are found, it will format them and forward them to the sovereign shard."
	app.Flags = []cli.Flag{
		logLevel,
		logSaveFile,
		disableAnsiColor,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = startNotifier

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startNotifier(ctx *cli.Context) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	err = initializeLogger(ctx)
	if err != nil {
		return err
	}

	var logFile closing.Closer
	withLogFile := ctx.GlobalBool(logSaveFile.Name)
	if withLogFile {
		logFile, err = createLogger()
		if err != nil {
			return err
		}
	}

	wsClient, err := factory.CreateWsSovereignNotifier(cfg)
	if err != nil {
		return fmt.Errorf("cannot create sovereign notifier, error: %w", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("starting ws client...")

	go wsClient.Start()

	<-interrupt
	log.Info("closing app at user's signal")

	err = wsClient.Close()
	log.LogIfError(err)

	if withLogFile {
		err = logFile.Close()
		log.LogIfError(err)
	}
	return nil
}

func loadConfig(filepath string) (config.Config, error) {
	cfg := config.Config{}
	err := core.LoadTomlFile(&cfg, filepath)

	log.Info("loaded config", "path", configPath)

	return cfg, err
}

func initializeLogger(ctx *cli.Context) error {
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err := logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return err
	}

	disableAnsi := ctx.GlobalBool(disableAnsiColor.Name)
	return removeANSIColorsForLoggerIfNeeded(disableAnsi)
}

func removeANSIColorsForLoggerIfNeeded(disableAnsi bool) error {
	if !disableAnsi {
		return nil
	}

	err := logger.RemoveLogObserver(os.Stdout)
	if err != nil {
		return err
	}

	return logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
}

func createLogger() (closing.Closer, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Error("error getting working directory when trying to create logger file", "error", err)
		workingDir = ""
	}

	argsLogger := file.ArgsFileLogging{
		WorkingDir:      workingDir,
		DefaultLogsPath: logsPath,
		LogFilePrefix:   logFilePrefix,
	}
	fileLogging, err := file.NewFileLogging(argsLogger)
	if err != nil {
		return nil, fmt.Errorf("%w creating log file", err)
	}

	err = fileLogging.ChangeFileLifeSpan(time.Second*time.Duration(logLifeSpanSec), logLifeSpanMb)
	if err != nil {
		return nil, err
	}

	return fileLogging, nil
}
