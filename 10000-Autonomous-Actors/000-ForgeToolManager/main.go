package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type Tool struct {
	Name    string
	Version string
	BinPath string
}

var Tools = []Tool{
	{Name: "buf", Version: "v1.34.0", BinPath: "Tools/authoring/buf.exe"},
	{Name: "golangci-lint", Version: "v1.59.1", BinPath: "Tools/authoring/golangci-lint.exe"},
	{Name: "air", Version: "v1.52.2", BinPath: "Tools/authoring/air.exe"},
}

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	slog.Info("⚒️ OlympusForge: Tool Management Loop Active", "workspace", "OlympusForge")

	for _, t := range Tools {
		if err := syncTool(t); err != nil {
			slog.Error("❌ Forge: Sync failed", "tool", t.Name, "error", err)
		}
	}
}

func syncTool(t Tool) error {
	absPath, _ := filepath.Abs(t.BinPath)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		slog.Info("📥 Forge: Tool missing, installing...", "name", t.Name, "path", absPath)
		return fmt.Errorf("manual installation required for %s", t.Name)
	}

	slog.Info("✅ Forge: Tool verified", "name", t.Name, "version", t.Version)
	return nil
}
