package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/srl-labs/ndk-greeter-go/greeter"
	"google.golang.org/grpc/metadata"
)

const (
	logTimeFormat = "2006-01-02 15:04:05 MST"
	version       = "0.1.0"
)

func main() {
	versionFlag := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
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

	app := greeter.NewApp(ctx, &logger)

	ctx = metadata.AppendToOutgoingContext(ctx, "agent_name", greeter.AppName)
	app.Start(ctx)
}
