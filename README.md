# OlympusForge

**The Authoritative Fleet Forge**

OlympusForge is the fleet's central manufacturing and build orchestrator. It provides a Go-native, high-performance pipeline for transforming user intent into deterministic system artifacts and managing the global toolchain.

## 🚀 Key Features

*   **Grammar-First Architecture**: Every workspace is defined by a formal grammar.
*   **jeBNF Optimization**: Hard-coded support for ordered choices (`/`), eliminating ambiguity.
*   **Go-Native Materializer**: One-click workspace bootstrapping from eBNF programs.
*   **Sovereign Toolchain Management**: Centralized registry and provisioning for fleet-wide binaries.

## 🛠️ Getting Started

### 1. Build the Forge Tooling
```bash
cd 00SDLC/OlympusForge
go build -o forge.exe ./provision.go
```

### 2. Validate a Grammar
```bash
gemaid-run jebnf-lint path/to/grammar.jebnf
```

### 3. Register a Tool
```bash
gemaid-run gemaid-finder -root . -rebuild
```

## 🏗️ Workspace Structure
OlympusForge follows the Sovereign Taxonomy (00000-90000), ensuring strict parity between AI reasoning and physical file structures.


## ⚖️ License
Proprietary - VelociKey LLC.
