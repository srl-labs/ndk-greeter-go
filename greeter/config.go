package greeter

import (
	"encoding/json"
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

// loadConfig loads configuration changes for greeter application.
// --8<-- [start:load-greeter-cfg].
func (a *App) loadConfig() {
	a.configState = &ConfigState{} // clear the configState
	if a.NDKAgent.Notifications.FullConfig != nil {
		err := json.Unmarshal(a.NDKAgent.Notifications.FullConfig, a.configState)
		if err != nil {
			a.logger.Error().Err(err).Msg("Failed to unmarshal config")
		}
	}
}

// --8<-- [end:load-greeter-cfg]

// processConfig processes the configuration received from the config notification stream
// and retrieves the uptime from the system.
// --8<-- [start:process-config].
func (a *App) processConfig() {
	if a.configState.Name == "" { // config is empty
		return
	}

	uptime, err := a.getUptime()
	if err != nil {
		a.logger.Info().Msgf("failed to get uptime: %v", err)
		return
	}

	// --8<-- [start:greeting-msg].
	a.configState.Greeting = "👋 Hi " + a.configState.Name +
		", SR Linux was last booted " + uptime + " ago!"
	// --8<-- [end:greeting-msg].
}

// --8<-- [end:process-config].
