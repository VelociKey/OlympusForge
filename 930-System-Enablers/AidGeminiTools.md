# AidGeminiTools: Sovereign Orchestrator Documentation

This document serves as the internal registry and operational guide for the **AidGeminiTools** pillar—a set of specialized Go-based orchestrators optimized for Agent-Mesh efficiency and fleet-wide consistency.

---

## 🛠 fleet-resetimportmodwork (ADR-016 Edition)
*Also known as or related to: **fleet-coord-sync***

### Overview
`fleet-resetimportmodwork` is the authoritative tool for Go module consolidation across the Olympus fleet. It enforces a **"One Module Per Workspace"** architecture (as defined in **ADR-016**) by centralizing module definitions and normalizing import paths to ensure O(1) high-performance resolution via the `MODULE_INDEX.jebnf`. In the `TOOLS_RATIONALE.md`, this functionality is primarily categorized under **fleet-coord-sync** for its role in automating `go.work` plumbing.

### Core Responsibilities
1.  **Workspace Discovery**: Identifies all fleet workspaces by scanning for `.git` roots and cross-referencing with `conductor/symbols/MODULE_INDEX.jebnf`.
2.  **Global Purge**: Recursively removes all nested `go.mod`, `go.sum`, and `go.work` files that are not at the root of an identified workspace.
3.  **Root Initialization**: Ensures every workspace has a valid `go.mod` (named after the workspace directory) and a localized `go.work` file.
4.  **Import Normalization**:
    *   Removes `github.com/VelociKey/` prefixes from all internal Go imports.
    *   Rewrites imports to point to the canonical `OlympusForge` relative paths for shared tools and externalized dependencies.
    *   Resolves workspace-relative imports to their flat-module names.
5.  **Master Integration**: Resets the root `go.work` file to include all discovered fleet workspaces, enabling seamless cross-module development.

### Usage
Run this tool from the fleet root (`C:/aAntigravitySpace`) to synchronize the environment after structural changes or module additions.

```powershell
# Preview changes (Recommended)
./00SDLC/OlympusForge/90000-Enablement-Labs/000-Tools/990-Execution-Hub/fleet-resetimportmodwork.exe -dryrun

# Apply consolidation
./00SDLC/OlympusForge/90000-Enablement-Labs/000-Tools/990-Execution-Hub/fleet-resetimportmodwork.exe
```

### Why We Use It
*   **Token Conservation**: By eliminating redundant `replace` directives and nested module overhead, we minimize the context required for Go builds and agent understanding.
*   **Architectural Integrity**: Enforces the ADR-016 standard, preventing "Import Drift" and ensuring all agents operate on a unified dependency graph.
*   **Performance**: Shifts the burden of module resolution from slow filesystem scans to high-performance indexed lookups.

---

## 📋 Other AidGeminiTools (Registry)

As defined in `TOOLS_RATIONALE.md`, the AidGeminiTools pillar includes:

*   **fleet-coord-sync** (`fleet-resetimportmodwork`): Automates the complex plumbing of `go.work` at O(1) efficiency.
*   **fleet-track**: Captures the "Semantic History" (the why) of our development, including economic labor metrics.
*   **fleet-impact**: Uses topological analysis to predict the "Blast Radius" of a code change.
*   **fleet-census**: Provides quantitative validation of our AI-assisted labor efficiency.
*   **fleet-chronicle**: Synthesizes Tracy Kidder style narratives to keep stakeholders aligned with the technical struggle.

*(Detailed documentation for additional tools to be expanded)*
