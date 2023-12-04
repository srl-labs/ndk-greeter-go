package greeter

import (
	"context"
	"encoding/json"
	"strings"

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

	// buffer aggregates the config notifications received.
	buffer []*ndk.ConfigNotification
	// receivedCh chan receives the value when the full config
	// is received by the stream client.
	receivedCh chan struct{}
}

// --8<-- [end:configstate-struct]

// receiveConfigNotifications receives a stream of configuration notifications
// buffer them in the configuration buffer and populates ConfigState struct of the App
// once the whole committed config is received.
// --8<-- [start:handle-cfg-notif].
func (a *App) receiveConfigNotifications(ctx context.Context) {
	bufFilledCh := make(chan struct{})

	go a.receiveAndBufferConfigNotifications(ctx, bufFilledCh)

	for {
		select {
		case <-ctx.Done():
			a.logger.Info().Msg("Context done, quitting configuration receive loop")
			return

		case <-bufFilledCh:
			a.logger.Info().Msg("Config notifications buffered, processing config")

			for _, cfg := range a.configState.buffer {
				a.handleGreeterConfig(ctx, cfg)
			}

			a.logger.Debug().Msg("Configuration has been read, clearing the config buffer")
			a.configState.buffer = make([]*ndk.ConfigNotification, 0)

			a.configState.receivedCh <- struct{}{}
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
		a.configState = &ConfigState{}

	default:
		a.logger.Info().Msgf("Handling create or update for .greeter config tree: %+v", cfg)

		err := json.Unmarshal([]byte(cfg.GetData().GetJson()), a.configState)
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
				Msgf("Storing config notification in buffer: %+v", cfgNotif)

			a.configState.buffer = append(a.configState.buffer, cfgNotif)
		}

		if cfgNotif.Key.JsPath == commitEndKeyPath && len(a.configState.buffer) > 0 {
			a.logger.Debug().
				Msgf("Received commit end notification: %+v", cfgNotif)

			doneCh <- struct{}{}
		}
	}
}

// --8<-- [end:buffer-cfg-notif]

func (a *App) processConfig(ctx context.Context) {
	if a.configState.Name == "" {
		a.logger.Info().Msg("No name configured, deleting state")

		return
	}

	uptime, err := a.getUptime(ctx)
	if err != nil {
		a.logger.Info().Msgf("failed to get uptime: %v", err)
		return
	}

	a.configState.Greeting = "👋 Hi " + a.configState.Name +
		", SR Linux was last booted at " + uptime
}

func (a *App) receiveAndBufferConfigNotifications(ctx context.Context, bufFilledCh chan struct{}) {
	configStream := a.StartConfigNotificationStream(ctx)

	for cfgStreamResp := range configStream {
		b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(cfgStreamResp)
		if err != nil {
			a.logger.Info().
				Msgf("Config notification Marshal failed: %+v", err)
			continue
		}

		a.logger.Info().
			Msgf("Received notifications:\n%s", b)

		a.bufferConfigNotifications(cfgStreamResp, bufFilledCh)
	}
}
