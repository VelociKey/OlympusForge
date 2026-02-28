# Sovereign Tool Rationale & SDLC Strategy

This document codifies the "Why" behind the Olympus toolchain. Every tool in our registry exists to satisfy one of our core mandates: **Sovereignty, Hermeticity, or Autonomy.**

---

## 🏗️ 1. ExternalTools (The Hermetic Pillar)
**Rationale**: Olympus must be a self-bootstrapping product. We do not rely on the customer's host OS for our engineering environment. We internalize third-party dependencies into the Forge to ensure deterministic builds and zero-drift deployments.

*   **Dagger**: The engine of our "Build Guardian." It ensures all artifacts are built in a clean-room environment.
*   **Ollama/Gemma3**: Our local intelligence core. We run models locally to maintain data sovereignty and avoid third-party inference costs.
*   **Go / Java / Flutter**: The foundational languages of our fleet. We version-lock these to ensure absolute symmetry across all workstations.
*   **GCloud / Firebase**: The bridge to our cloud substrate. Internalized to ensure emulators are always available for local-first development.
*   **LanceDB / DuckDB**: Specialized data engines for agent memory and high-performance local analytics.

## 🛠️ 2. SupportTools (The Utility Pillar)
**Rationale**: These tools provide the "Standard Gauge" for our data and security. They are the low-level enablers that allow high-level orchestrators to function.

*   **Buf / protoc-gen-***: Ensures that our communication contracts (ConnectRPC) are strictly typed and generated across Go and Dart.
*   **Trivy / Govulncheck / Cosign**: The "Verify, Seal, and Scan" gauntlet. Every entry into the fleet is cryptographically and logically validated.
*   **JQ / Air / Delve**: Developer-experience tools optimized for the Sovereign workflow (JSON processing, live-reload, debugging).
*   **Promptfoo / Genkit**: Specialized frameworks for testing and building the reasoning flows of our agents.

## 🤖 3. AidGeminiTools (The Agent Pillar)
**Rationale**: These are the "Super-Tools" used by Gemini-CLI and Antigravity. They transform the fleet from a set of files into a living, graph-aware intelligence.

*   **fleet-coord-sync**: Automates the complex plumbing of `go.work` at O(1) efficiency.
*   **fleet-track**: Captures the "Semantic History" (the why) of our development, including economic labor metrics.
*   **fleet-impact**: Uses topological analysis to predict the "Blast Radius" of a code change.
*   **fleet-census**: Provides quantitative validation of our AI-assisted labor efficiency.
*   **fleet-chronicle**: Synthesizes Tracy Kidder style narratives to keep stakeholders aligned with the technical struggle.

## 📦 4. InternalTools (The Product Pillar)
**Rationale**: These are the actual binaries built from our source code. They represent the "Deliverable Value" of the Olympus project.

*   **GCP Managers**: Autonomous actors that manage cloud infrastructure.
*   **InteractionSurface (Vision)**: The high-fidelity UI for human-agent interaction.

## 🧠 5. AgentTools (The Runtime Pillar)
**Rationale**: To maintain sovereignty, our agents must have functional runtimes that are independent of the SDLC tools. 

*   **George / Jules / Gemini-CLI**: These represent the specific agent personas. We separate their functional code from their management logic to allow them to evolve independently.

---
*Created by the Conductor Agent-Mesh (Feb 27, 2026)*
