// --8<-- [start:pkg-greeter]
package greeter

// --8<-- [end:pkg-greeter]

import (
	"context"
	"time"

	"github.com/openconfig/gnmic/pkg/api"
	"github.com/rs/zerolog"
	"github.com/srl-labs/bond"
)

// --8<-- [start:pkg-greeter-const].
const (
	AppName = "greeter"
	AppRoot = "/" + AppName
)

// --8<-- [end:pkg-greeter-const]

// App is the greeter application struct.
// --8<-- [start:app-struct].
type App struct {
	Name string
	// configState holds the application configuration and state.
	configState *ConfigState
	logger      *zerolog.Logger
	NDKAgent    *bond.Agent
}

// --8<-- [end:app-struct]

// NewApp creates a new Greeter App instance and connects to NDK socket.
// It also creates the NDK service clients and registers the agent with NDK.
// --8<-- [start:new-app].
func New(logger *zerolog.Logger, agent *bond.Agent) *App {
	return &App{
		Name:     AppName,
		NDKAgent: agent,

		configState: &ConfigState{},
		logger:      logger,
	}
}

// --8<-- [end:new-app]

// Start starts the application.
// --8<-- [start:app-start].
func (a *App) Start(ctx context.Context) {
	for {
		select {
		case <-a.NDKAgent.Notifications.FullConfigReceived:
			a.logger.Info().Msg("Received full config")

			a.loadConfig()

			a.processConfig()

			a.updateState()

		case <-ctx.Done():
			return
		}
	}
}

// --8<-- [end:app-start]

// getUpTime retrieves the uptime from the system using gNMI.
// --8<-- [start:get-uptime].
func (a *App) getUptime() (string, error) {
	a.logger.Info().Msg("Fetching SR Linux last-booted time value")

	// create a GetRequest
	getReq, err := bond.NewGetRequest("/system/information/last-booted", api.EncodingPROTO())
	if err != nil {
		return "", err
	}

	getResp, err := a.NDKAgent.GetWithGNMI(getReq)
	if err != nil {
		return "", err
	}

	a.logger.Info().Msgf("GetResponse: %+v", getResp)

	bootTimeStr := getResp.GetNotification()[0].GetUpdate()[0].GetVal().GetStringVal()

	bootTime, err := time.Parse(time.RFC3339Nano, bootTimeStr)
	if err != nil {
		return "", err
	}

	currentTime := time.Now()
	uptime := currentTime.Sub(bootTime).Round(time.Second)

	return uptime.String(), nil
}

// --8<-- [end:get-uptime].
