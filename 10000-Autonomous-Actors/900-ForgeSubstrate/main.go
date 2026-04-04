package main

import (
	"log/slog"
	"os"
	"os/exec"
)

func main() {
	slog.Info("Olympus Forge: Substrate Provisioner starting")
	slog.Info("Forging the workstation-local GCP infrastructure...")

	composePath := "../../60000-Information-Storage/900-Emulators/podman-compose.yaml"

	cmd := exec.Command("podman-compose", "-f", composePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	slog.Info("Starting Substrate via Podman Compose...")
	if err := cmd.Run(); err != nil {
		slog.Error("Failed to start substrate", "error", err)
		os.Exit(1)
	}

	slog.Info("Substrate Active",
		"PubSub", "localhost:8085",
		"Datastore", "localhost:8090",
		"Bigtable", "localhost:8086",
		"Firestore", "localhost:8081",
		"Spanner_gRPC", "localhost:9010",
		"Spanner_HTTP", "localhost:9020",
		"Storage", "localhost:4443",
		"SecretManager_Vault", "localhost:8200",
		"VertexAI_Ollama", "localhost:11434",
	)
}
