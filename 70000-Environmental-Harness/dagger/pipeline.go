package main

import (
	"context"
	"dagger/olympusforge/internal/dagger"
)

// BuildGemAid orchestrates the build of the OlympusGemAid rust binary.
func (m *Olympusforge) BuildGemAid(ctx context.Context, src *dagger.Directory) *dagger.File {
	// The path is relative to the provided source directory.
	// We assume the standard fleet structure where src is the repo root.
	projectPath := "00SDLC/OlympusGemAid"
	binaryName := "gem_aid.exe"

	return m.BuildRust(ctx, src, projectPath, binaryName)
}
