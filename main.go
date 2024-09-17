// Main package.
// --8<-- [start:pkg-main]
package main

// --8<-- [end:pkg-main]

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	syslog "github.com/RackSec/srslog"

	"github.com/rs/zerolog"
	"github.com/srl-labs/bond"
	"github.com/srl-labs/ndk-greeter-go/greeter"
	"gopkg.in/natefinch/lumberjack.v2"
)

// --8<-- [start:pkg-main-vars].
var (
	version = "0.0.0"
	commit  = ""
)

// --8<-- [end:pkg-main-vars]

// Main entry point for the application.
// --8<-- [start:main].
func main() {
	versionFlag := flag.Bool("version", false, "print the version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Println(version + "-" + commit)
		os.Exit(0)
	}

	logger := setupLogger()

	// --8<-- [start:metadata]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// --8<-- [end:metadata]

	// --8<-- [start:main-init-bond-agent]
	opts := []bond.Option{
		bond.WithLogger(&logger),
		bond.WithContext(ctx, cancel),
		bond.WithAppRootPath(greeter.AppRoot),
	}

	agent, errs := bond.NewAgent(greeter.AppName, opts...)
	for _, err := range errs {
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create agent")
		}
	}

	err := agent.Start()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to start agent")
	}
	// --8<-- [end:main-init-bond-agent]

	// --8<-- [end:main-init-app]
	app := greeter.New(&logger, agent)
	app.Start(ctx)
	// --8<-- [end:main-init-app]
}

// --8<-- [end:main]

// setupLogger creates a logger instance.
// --8<-- [start:setup-logger].
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

	// --8<-- [start:syslog-logger]
	var zsyslog zerolog.SyslogWriter
	zsyslog, err = syslog.Dial("", "", syslog.LOG_INFO|syslog.LOG_LOCAL7, "ndk-greeter-go")
	if err != nil {
		panic(err)
	}
	// --8<-- [end:syslog-logger]

	writers = append(writers, fileLogger, zsyslog)

	mw := io.MultiWriter(writers...)

	return zerolog.New(mw).With().Caller().Timestamp().Logger()
}

// --8<-- [end:setup-logger]
