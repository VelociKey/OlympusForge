# Comprehensive External Tool Automation Audit (CYC-095)

## Status: COMPLETED
**Date**: 2026-03-01

This manifest provides the definitive set of external tools required by the Sovereign Fleet and their associated automation/ingest paths.

### 1. System Prerequisite Engine (Winget / NPM)
These tools are managed via system package managers and verified by `fleet-bootstrap`.

| Tool | Winget/NPM ID | Verification Command |
| :--- | :--- | :--- |
| **Podman Desktop** | `RedHat.Podman-Desktop` | `podman --version` |
| **Flutter SDK** | `Google.Flutter` | `flutter --version` |
| **Firebase CLI** | `firebase-tools` (NPM) | `firebase --version` |
| **Google Cloud CLI** | `Google.CloudSDK` | `gcloud --version` |
| **Go Runtime** | `GoLang.Go` | `go version` |
| **Java JDK (25)** | `Oracle.JDK.25` | `java -version` |
| **Kotlin** | `JetBrains.Kotlin` | `kotlinc -version` |
| **Gradle** | `Gradle.Gradle` | `gradle -v` |
| **Git CLI** | `Git.Git` | `git --version` |
| **GitHub CLI** | `GitHub.cli` | `gh --version` |
| **OpenTofu** | `OpenTofu.OpenTofu` | `tofu --version` |
| **DuckDB** | `DuckDB.DuckDB` | `duckdb --version` |
| **JQ** | `jqlang.jq` | `jq --version` |
| **Trivy** | `AquaSecurity.Trivy` | `trivy --version` |

### 2. Language-Specific Tooling (Go Install / NPM)
These tools are ingested directly into the Forge via language-specific commands.

| Tool | Source Path | Ingest Target |
| :--- | :--- | :--- |
| **Buf** | `github.com/bufbuild/buf/cmd/buf` | 81000/buf.exe |
| **Cosign** | `github.com/sigstore/cosign/v2/cmd/cosign` | 81000/cosign.exe |
| **Nuclei** | `github.com/projectdiscovery/nuclei/v3/cmd/nuclei` | 81000/nuclei.exe |
| **Gosec** | `github.com/securego/gosec/v2/cmd/gosec` | 81000/gosec.exe |
| **Air** | `github.com/air-verse/air` | 81000/air.exe |
| **Delve** | `github.com/go-delve/delve/cmd/dlv` | 81000/dlv.exe |
| **Promptfoo** | `promptfoo` (NPM) | 81000/promptfoo.cmd |
| **Genkit** | `genkit` (NPM) | 81000/genkit.cmd |

### 3. Manual / Script-Based Gaps
These tools require special handling as they lack standard CLI package definitions.

- **Dagger CLI**: Downloaded from `dagger.io`. Needs automated wrapper.
- **Otelcol**: OpenTelemetry collector binary. Downloaded from GitHub releases.
- **Shorebird**: Flutter code-push. Installed via shell script.
- **Ollama Models**: Weights are downloaded via `ollama pull`.

## Verification Summary
Total Tools Audited: 29
Automated Paths: 25
Manual Gaps: 4
