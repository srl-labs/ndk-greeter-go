package greeter

import (
	"encoding/json"
)

// updateState updates the state of the application.
// --8<-- [start:update-state].
func (a *App) updateState() {
	jsData, err := json.Marshal(a.configState)
	if err != nil {
		a.logger.Info().Msgf("failed to marshal json data: %v", err)
		return
	}

	a.NDKAgent.UpdateState(AppRoot, string(jsData))
}

// --8<-- [end:update-state].
