package main

import "context"
import "dagger/olympusforge/internal/dagger"

type OlympusForge struct{}

func (m *OlympusForge) HelloWorld(ctx context.Context) string { return "Hello from OlympusForge!" }

func main() { dagger.Serve() }
