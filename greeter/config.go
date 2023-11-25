package greeter

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/nokia/srlinux-ndk-go/ndk"
)

const (
	commitEndKeyPath = ".commit.end"
	greeterKeyPath   = ".greeter"
)

// ConfigState holds the application configuration and state.
type ConfigState struct {
	// Name is the name to use in the greeting.
	Name string `json:"name"`
	// Greeting is the greeting message to be displayed.
	Greeting string `json:"greeting,omitempty"`
}

func (a *App) handleConfigNotifications(ctx context.Context, notifStreamResp *ndk.NotificationStreamResponse) {
	buf := a.bufferConfigNotifications(notifStreamResp)

	// commit end notification received
	// process config buffer
	for _, cfg := range buf {
		switch cfg.Key.JsPath {
		case greeterKeyPath:
			a.handleGreeterConfig(ctx, cfg)
		}
	}
}

// handleGreeterConfig handles configuration changes for greeter application.
func (a *App) handleGreeterConfig(ctx context.Context, cfg *ndk.ConfigNotification) {
	switch cfg.GetOp() {
	case ndk.SdkMgrOperation_Create, ndk.SdkMgrOperation_Update:
		// upon sr linux boot, the first config notification contains empty data
		// we skip it, as we do not carry out any action on empty config.
		if strings.TrimSpace(cfg.GetData().GetJson()) == "{\n}" {
			a.logger.Info().Msgf("Empty config data for create/update operation for key %q", cfg.GetKey().GetJsPath())
			return
		}

		a.logger.Info().Msgf("Handling create or update for .greeter config tree: %+v", cfg)
		a.handleGreeterCreateOrUpdate(ctx, cfg.GetData())
	}

	a.updateGreeterState(ctx)
}

func (a *App) handleGreeterCreateOrUpdate(ctx context.Context, data *ndk.ConfigData) {
	// read the config into the application config struct
	err := json.Unmarshal([]byte(data.GetJson()), a.ConfigState)
	if err != nil {
		a.logger.Error().Msgf("failed to unmarshal path %q config %+v", ".greeter", data)
		return
	}
}

// bufferConfigNotifications buffers the configuration notifications received
// from the config notification stream before commit end notification is received.
func (a *App) bufferConfigNotifications(notifStreamResp *ndk.NotificationStreamResponse) []*ndk.ConfigNotification {
	notifs := notifStreamResp.GetNotification()
	// buf holds the configuration notifications received before commit end.
	buf := make([]*ndk.ConfigNotification, 0, len(notifs))

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
			a.logger.Info().
				Msgf("Storing config notification in buffer:%+v", cfgNotif)

			buf = append(buf, cfgNotif)
		}
	}

	return buf
}
