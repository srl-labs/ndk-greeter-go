package greeter

import (
	"context"
	"encoding/json"
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

	// receivedCh chan receives the value when the full config
	// is received by the stream client.
	receivedCh chan struct{}
}

// --8<-- [end:configstate-struct]

// receiveConfigNotifications receives a stream of configuration notifications
// buffer them in the configuration buffer and populates ConfigState struct of the App
// once the whole committed config is received.
// --8<-- [start:rcv-cfg-notif].
func (a *App) receiveConfigNotifications(ctx context.Context) {
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

		a.handleConfigNotifications(cfgStreamResp)
	}
}

// --8<-- [end:rcv-cfg-notif]

// handleGreeterConfig handles configuration changes for greeter application.
// --8<-- [start:handle-greeter-cfg].
func (a *App) handleGreeterConfig(cfg *ndk.ConfigNotification) {
	switch {
	case a.isEmptyObject(cfg.GetData().GetJson()):
		m := sync.Mutex{}
		m.Lock()

		a.logger.Info().Msgf("Handling deletion of the .greeter config tree: %+v", cfg)

		a.configState.Name = ""
		a.configState.Greeting = ""

		m.Unlock()

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

// handleConfigNotifications buffers the configuration notifications received
// from the config notification stream before commit end notification is received.
// --8<-- [start:buffer-cfg-notif].
func (a *App) handleConfigNotifications(
	notifStreamResp *ndk.NotificationStreamResponse,
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
				Msgf("Handling config notification: %+v", cfgNotif)

			a.handleGreeterConfig(cfgNotif)
		}

		if cfgNotif.Key.JsPath == commitEndKeyPath &&
			!a.isCommitSeqZero(cfgNotif.GetData().GetJson()) {
			a.logger.Debug().
				Msgf("Received commit end notification: %+v", cfgNotif)

			a.configState.receivedCh <- struct{}{}
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

type CommitSeq struct {
	CommitSeq int `json:"commit_seq"`
}

// isCommitSeqZero checks if the commit sequence passed in the jsonStr is zero.
func (a *App) isCommitSeqZero(jsonStr string) bool {
	var commitSeq CommitSeq

	err := json.Unmarshal([]byte(jsonStr), &commitSeq)
	if err != nil {
		a.logger.Error().Msgf("failed to unmarshal json: %s", err)
		return false
	}

	return commitSeq.CommitSeq == 0
}

// isEmptyObject checks if the jsonStr is an empty object.
func (a *App) isEmptyObject(jsonStr string) bool {
	var obj map[string]any

	err := json.Unmarshal([]byte(jsonStr), &obj)
	if err != nil {
		a.logger.Error().Msgf("failed to unmarshal json: %s", err)
		return false
	}

	if len(obj) == 0 {
		return true
	}

	return false
}
