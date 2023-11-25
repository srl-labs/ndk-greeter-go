package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/srl-labs/ndk-greeter-go/greeter"
	"google.golang.org/grpc/metadata"
)

const (
	logTimeFormat = "2006-01-02 15:04:05 MST"
)

var (
	version = "0.0.0"
	commit  = ""
)

func main() {
	versionFlag := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version + "-" + commit)
		os.Exit(0)
	}

	// set logger parameters
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: logTimeFormat,
		NoColor:    true,
	}).With().Timestamp().Logger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "agent_name", greeter.AppName)

	app := greeter.NewApp(ctx, &logger)

	exitHandler(cancel)

	app.Start(ctx)
}

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
