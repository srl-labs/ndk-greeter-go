package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/srl-labs/ndk-greeter-go/greeter"
	"google.golang.org/grpc/metadata"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	version = "0.0.0"
	commit  = ""
)

// Main entry point for the application.
// --8<-- [start:main]
func main() {
	versionFlag := flag.Bool("version", false, "print the version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Println(version + "-" + commit)
		os.Exit(0)
	}

	logger := setupLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "agent_name", greeter.AppName)

	app := greeter.NewApp(ctx, &logger)

	exitHandler(cancel)

	app.Start(ctx)
}

// --8<-- [end:main]

// ExitHandler cancels the main context when interrupt or term signals are sent.
func exitHandler(cancel context.CancelFunc) {
	// handle CTRL-C signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig

		cancel()
	}()
}

func setupLogger() zerolog.Logger {
	var writers []io.Writer

	// the lab creates an empty file to indicate
	// that we run in dev mode. If file exists, we
	// log to console as well.
	_, err := os.Stat("/tmp/.ndk-dev-mode")
	if err == nil {
		const logTimeFormat = "2006-01-02 15:04:05 MST"

		consoleLogger := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: logTimeFormat,
			NoColor:    true,
		}

		writers = append(writers, consoleLogger)
	}

	// A lumberjack logger with rotation settings.
	fileLogger := &lumberjack.Logger{
		Filename:   "/var/log/greeter/greeter.log",
		MaxSize:    2, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	}

	writers = append(writers, fileLogger)

	mw := io.MultiWriter(writers...)

	return zerolog.New(mw).With().Timestamp().Logger()
}
