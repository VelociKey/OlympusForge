# Sovereign Tool Rationale & SDLC Strategy

This document codifies the "Why" behind the Olympus toolchain. Every tool in our registry exists to satisfy one of our core mandates: **Sovereignty, Hermeticity, or Autonomy.**

---

## 🏛️ The Sovereign Trinity Architecture (v2.2)

Following the fleet re-organization, we have categorized all assets into three functional layers:

### **Decoupling Identity from Location**
The fleet's organization into **Planes (P0, P1, P2)** defines governance and insulation, but **Discovery** is organizational-agnostic. Agents request tools by their **Canonical ID** (e.g., `trivy`), and the **Authoritative Index** handles the resolution to a physical path transparently.

### **1. The Substrate (P0 - Foundation)**
*   **Purpose**: Hermetic external tools and libraries we depend upon.
*   **Categories**: External libraries (Go, Java), Security (Trivy), Substrate management (Podman).

### **2. The Engine (P2 - Intelligence)**
*   **Purpose**: The internal power of the fleet. Separated by **Role**:
    *   **Forge Workers**: Tools that "produce other things" (Materializers, Scaffolders).
    *   **Sovereign Intermediaries (GemAids)**: The runtime layer (Finders, Orchestrators, SCM, Sanitization).
    - **Production Agents**: Autonomous personas like Gemini-CLI that orchestrate the Workers and Intermediaries.

### **3. The Silos (P1 - Products)**
*   **Purpose**: Revenue-generating deliverables. (e.g., 00SDLC, 30INFR).
*   **Insulation Mandate**: Each product area is a separate silo. Subdirectories are strictly insulated from each other to prevent dependency drift and maximize deliverable autonomy.

---

## 🏗️ 1. ExternalTools (The Hermetic Pillar - P0)
**Rationale**: We do not rely on the customer's host OS. This includes third-party languages and pre-existing platform agents.

*   **Go / Java / Flutter**: Foundational languages. (P0)
*   **External Agents**: Gemini-CLI, Jules, Hermes. These are the external cognitive powers we bring into the fleet.
*   **Trivy / Govulncheck**: Security gauntlet. (P0)

## 🛠️ 2. Helper Tools (The Utility Pillar - P2)
**Rationale**: Static builders and platform abstractions.

*   **OlympusFabric**: Formal grammar manufacturing.
*   **OlympusGemAid**: The **Sovereign Runtime Environment**.
    - **gemaid-finder**: Unified discovery (Discovery + Search).
    - **gemaid-run**: Execution orchestrator (@RUN).
    - **gemaid-scm**: Platform-agnostic Git porcelain (Commit/Publish).
    - **gemaid-sanitize**: Automatic fleet-wide Version and Import synching.

## 🤖 3. Production Agents (The Agent Pillar - P2)
**Rationale**: Autonomous actors we build ourselves to automate the SDLC and help with code production.

*   **Antigravity**: The cross-workspace orchestrator.
*   **Fleet-Native Mesh**: Agents designed specifically for Olympus production.

## 📦 4. Product Offerings (The Silo Pillar - P1)
**Rationale**: Revenue-generating deliverables organized in insulated zones.

*   **00SDLC**: The Software Development Life Cycle Platform.
*   **30INFR**: Infrastructure-as-a-Product.

## 🚀 5. Silo-to-Tool Promotion
**Rationale**: Some products created within silos are tools that must be used by other parts of the fleet. We promote these to the "Engine" (P2) or "Substrate" (P0) to ensure utility without breaking source-level insulation.

*   **Rule**: The product source stays in its Silo (P1).
*   **Action**: The *built artifact* (binary/library) is injected directly into the **Unified Warehouse** (`@BIN`).
*   **Result**: Other workspaces consume the *tool* from the warehouse, never the *source*, maintaining strict P1 insulation.

## ⚙️ 6. Unified Build Injection
**Rationale**: To prevent environment drift, we maintain a strict separation between **Source Code** and **Executable Binaries**.

*   **Global Mandate**: No workspace should contain its own `bin` directory or locals artifacts.
*   **Go Protocol**: Build commands must use `-o @BIN/toolname` to inject artifacts into the central locker.
*   **Workspace Integrity**: The `go.work` and `go.mod` files govern the *logical* relationships, while the Warehouse governs the *physical* execution.

---
*Updated for Unified Warehouse Governance (March 2026)*
