package main

import (
	"log/slog"
	"os"
	"os/exec"
)

type Tool struct {
	Name    string
	Version string
	BinPath string
}

// Tools managed by the Forge Tool Manager.
// Paths are relative to the root of the workspace (OlympusForge).
var Tools = []Tool{
	{Name: "buf", Version: "v1.50.0", BinPath: "../../81000-Toolchain-External/bufbuild/bin/buf.exe"},
	{Name: "cosign", Version: "v2.4.1", BinPath: "../../81000-Toolchain-External/sigstore/bin/cosign.exe"},
	{Name: "protoc-gen-go", Version: "v1.36.1", BinPath: "../../81000-Toolchain-External/google/bin/protoc-gen-go.exe"},
	{Name: "protoc-gen-connect-go", Version: "v1.18.1", BinPath: "../../81000-Toolchain-External/connectrpc/bin/protoc-gen-connect-go.exe"},
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
	// Paths are relative to the workspace root
	if _, err := os.Stat(t.BinPath); os.IsNotExist(err) {
		slog.Info("📥 Forge: Tool missing, invoking Fleet Bootstrap...", "name", t.Name, "path", t.BinPath)
		
		// Run the central bootstrap tool
		bootstrapPath := "../../82000-Toolchain-Fleet/fleet-bootstrap/main.go" 
		cmd := exec.Command("go", "run", bootstrapPath, "--install", "--update")
		cmd.Dir = "C:/aAntigravitySpace" // Fleet root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	slog.Info("✅ Forge: Tool verified", "name", t.Name, "version", t.Version)
	return nil
}
