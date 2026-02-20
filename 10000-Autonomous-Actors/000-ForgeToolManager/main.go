package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type Tool struct {
	Name    string
	Version string
	BinPath string
}

var Tools = []Tool{
	{Name: "buf", Version: "v1.34.0", BinPath: "000-Tools/authoring/buf.exe"},
	{Name: "golangci-lint", Version: "v1.59.1", BinPath: "000-Tools/authoring/golangci-lint.exe"},
	{Name: "air", Version: "v1.52.2", BinPath: "000-Tools/authoring/air.exe"},
	{Name: "dagger", Version: "v0.19.11", BinPath: "000-Tools/infrastructure/dagger.exe"},
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
	// Root is relative to workspace
	absPath, _ := filepath.Abs(filepath.Join("../../90000-Enablement-Labs", t.BinPath))
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		slog.Info("📥 Forge: Tool missing, invoking Master Provisioner...", "name", t.Name, "path", absPath)
		
		// Run the central provisioner
		provisionerPath := "C:/aAntigravitySpace/Olympus2/90000-Enablement-Labs/910-provisioner/main.go"
		cmd := exec.Command("go", "run", provisionerPath)
		cmd.Dir = "C:/aAntigravitySpace" // Fleet root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	slog.Info("✅ Forge: Tool verified", "name", t.Name, "version", t.Version)
	return nil
}
