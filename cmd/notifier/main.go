package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

var (
	log = logger.GetOrCreate("mx-chain-sovereign-notifier")
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
	cfg, err := loadMainConfig("config/config.toml")
	if err != nil {
		return err
	}

	fileLogging, err := initializeLogger(ctx, cfg)
	if err != nil {
		return err
	}

	//wsClient, err := factory.CreateWsIndexer(cfg, nil)
	//if err != nil {
	//	log.Error("cannot create ws indexer", "error", err)
	//}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	//go wsClient.Start()

	<-interrupt
	log.Info("closing app at user's signal")
	//wsClient.Close()
	if !check.IfNilReflect(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}
	return nil
}

func loadMainConfig(filepath string) (config.Config, error) {
	//cfg := config.Config{}
	//err := core.LoadTomlFile(&cfg, filepath)

	tomlBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config.Config{}, err
	}

	var cfg config.Config
	err = toml.Unmarshal(tomlBytes, &cfg)
	if err != nil {
		return config.Config{}, err
	}

	log.Info("cfg", "addr", cfg.SubscribedAddresses)

	return cfg, err
}

func initializeLogger(ctx *cli.Context, cfg config.Config) (closing.Closer, error) {
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err := logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return nil, err
	}

	withLogFile := ctx.GlobalBool(logSaveFile.Name)
	if !withLogFile {
		return nil, nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		log.LogIfError(err)
		workingDir = ""
	}

	fileLogging, err := file.NewFileLogging(file.ArgsFileLogging{
		WorkingDir:      workingDir,
		DefaultLogsPath: "logs",
		LogFilePrefix:   "sovereign-notifier",
	})
	if err != nil {
		return nil, fmt.Errorf("%w creating a log file", err)
	}

	err = fileLogging.ChangeFileLifeSpan(
		time.Second*time.Duration(432000),
		uint64(1024),
	)
	if err != nil {
		return nil, err
	}

	disableAnsi := ctx.GlobalBool(disableAnsiColor.Name)
	err = removeANSIColorsForLoggerIfNeeded(disableAnsi)
	if err != nil {
		return nil, err
	}

	return fileLogging, nil
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
