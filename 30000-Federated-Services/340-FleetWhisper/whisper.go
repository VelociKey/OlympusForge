package fleetwhisper

import (
	"context"
	"log/slog"

	whisperv1 "olympus.fleet/00SDLC/OlympusGrammar/gen/v1/whisper/whisperv1connect"

	"olympus.fleet/ext/connectrpc/connect-go"
)

// FleetWhisper is the forge-side bridge to the inter-agent event mesh.
type FleetWhisper struct {
	client whisperv1.WhisperBusServiceClient
}

func NewFleetWhisper(httpClient connect.HTTPClient, baseURL string) *FleetWhisper {
	// Implementation would initialize the ConnectRPC client
	return &FleetWhisper{}
}

// BroadcastSignal emits an infrastructure event to the fleet.
func (w *FleetWhisper) BroadcastSignal(ctx context.Context, topic, payload string) error {
	slog.Info("Broadcasting Fleet Signal", "topic", topic)

	// Real implementation would call Emit on the WhisperBus
	return nil
}
