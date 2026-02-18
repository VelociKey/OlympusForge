package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Tool struct {
	Name       string
	Type       string // "bin", "npm", "external", "model"
	Path       string
	InstallCmd string
}

var ToolStack = []Tool{
	{Name: "Olympus2/buf", Type: "Olympus2/bin", Path: "Olympus2/C0400-Artifacts/Tools/authoring/buf.exe"},
	{Name: "gcloud", Type: "external", InstallCmd: "gcloud components update --quiet"},
	{Name: "firebase", Type: "npm", InstallCmd: "npm install -g firebase-tools"},
	{Name: "ollama-models", Type: "model", InstallCmd: "ollama pull gemma2:2b"},
}

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	slog.Info("⚒️ ForgeManager: High-Level Platform Sync Active")

	for _, t := range ToolStack {
		if err := syncPlatformTool(t); err != nil {
			slog.Error("❌ Forge: Sync failed", "tool", t.Name, "error", err)
		}
	}
}

func syncPlatformTool(t Tool) error {
	switch t.Type {
	case "bin":
		absPath, _ := filepath.Abs(filepath.Join("..", "..", t.Path))
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("binary missing at %s", absPath)
		}
		slog.Info("✅ Forge: Binary verified", "name", t.Name)
	case "external", "npm", "model":
		slog.Info("🔄 Forge: Syncing platform component", "name", t.Name, "cmd", t.InstallCmd)

		parts := strings.Fields(t.InstallCmd)
		if len(parts) == 0 {
			return nil
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command execution failed: %w", err)
		}
		slog.Info("✅ Forge: Component updated", "name", t.Name)
	}
	return nil
}
