package greeter

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/nokia/srlinux-ndk-go/ndk"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	commitEndKeyPath = ".commit.end"
	greeterKeyPath   = ".greeter"
)

// ConfigState holds the application configuration and state.
// --8<-- [start:configstate-struct].
type ConfigState struct {
	// Name is the name to use in the greeting.
	Name string `json:"name,omitempty"`
	// Greeting is the greeting message to be displayed.
	Greeting string `json:"greeting,omitempty"`
}

// --8<-- [end:configstate-struct]

// receiveConfigNotifications handles the configuration notifications received.
// --8<-- [start:handle-cfg-notif].
func (a *App) receiveConfigNotifications(ctx context.Context) {
	configStream := a.StartConfigNotificationStream(ctx)

	bufDone := make(chan struct{})

	go func() {
		for cfgStreamResp := range configStream {
			b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(cfgStreamResp)
			if err != nil {
				a.logger.Info().
					Msgf("Config notification Marshal failed: %+v", err)
				continue
			}

			a.logger.Info().
				Msgf("Received notifications:\n%s", b)

			a.bufferConfigNotifications(cfgStreamResp, bufDone)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info().Msg("Context done, exiting config notification handler")
			return

		case <-bufDone:
			a.logger.Info().Msg("Buffer done, processing config")

			for _, cfg := range a.ConfigBuffer {
				a.handleGreeterConfig(ctx, cfg)
			}
			// clear the config buffer
			a.logger.Info().Msg("Clearing config buffer")
			a.ConfigBuffer = make([]*ndk.ConfigNotification, 0)
			a.ConfigReceived <- struct{}{}
		}
	}
}

// --8<-- [end:handle-cfg-notif]

// handleGreeterConfig handles configuration changes for greeter application.
// --8<-- [start:handle-greeter-cfg].
func (a *App) handleGreeterConfig(ctx context.Context, cfg *ndk.ConfigNotification) {
	switch {
	case strings.TrimSpace(cfg.GetData().GetJson()) == "{\n}":
		a.logger.Info().Msgf("Handling deletion of the .greeter config tree: %+v", cfg)
		a.ConfigState = &ConfigState{}

	default:
		a.logger.Info().Msgf("Handling create or update for .greeter config tree: %+v", cfg)

		err := json.Unmarshal([]byte(cfg.GetData().GetJson()), a.ConfigState)
		if err != nil {
			a.logger.Error().Msgf("failed to unmarshal path %q config %+v", ".greeter", cfg.GetData())
			return
		}
	}
}

// --8<-- [end:handle-greeter-cfg]

// bufferConfigNotifications buffers the configuration notifications received
// from the config notification stream before commit end notification is received.
// --8<-- [start:buffer-cfg-notif].
func (a *App) bufferConfigNotifications(
	notifStreamResp *ndk.NotificationStreamResponse, doneCh chan struct{},
) {
	notifs := notifStreamResp.GetNotification()

	for _, n := range notifs {
		cfgNotif := n.GetConfig()
		if cfgNotif == nil {
			a.logger.Info().
				Msgf("Empty configuration notification:%+v", n)
			continue
		}

		// do not include commit end notification in the buffer
		// as it is just an indication that the config is passed in full.
		if cfgNotif.Key.JsPath != commitEndKeyPath {
			a.logger.Debug().
				Msgf("Storing config notification in buffer:%+v", cfgNotif)

			a.addToConfigBuffer(cfgNotif)
		}

		if cfgNotif.Key.JsPath == commitEndKeyPath && len(a.ConfigBuffer) > 0 {
			a.logger.Debug().
				Msgf("Received commit end notification:%+v", cfgNotif)

			doneCh <- struct{}{}
		}
	}
}

// --8<-- [end:buffer-cfg-notif]

func (a *App) processConfig(ctx context.Context) {
	if a.ConfigState.Name == "" {
		a.logger.Info().Msg("No name configured, deleting state")

		return
	}

	uptime, err := a.getUptime(ctx)
	if err != nil {
		a.logger.Info().Msgf("failed to get uptime: %v", err)
		return
	}

	a.ConfigState.Greeting = "👋 Hello " + a.ConfigState.Name +
		", SR Linux was last booted at " + uptime
}

func (a *App) addToConfigBuffer(cfg *ndk.ConfigNotification) {
	var mutex sync.Mutex

	mutex.Lock()
	defer mutex.Unlock()

	a.ConfigBuffer = append(a.ConfigBuffer, cfg)
}
