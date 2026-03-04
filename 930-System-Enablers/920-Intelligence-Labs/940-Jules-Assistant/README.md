# Jules: Asynchronous Agentic Coding Assistant

Jules is the fleet's specialized engine for asynchronous background operations, primarily focused on test generation, deep codebase analysis, and long-running refactoring campaigns.

## Strategic Role

While the primary agent handles interactive sessions, Jules operates in the background to:
- Generate comprehensive test suites (Unit, Integration, E2E).
- Perform deep-dive structural audits.
- Execute fleet-wide technical debt remediation.

## Installation

Run the installation script to add Jules to your environment:
```powershell
./install.ps1
```

## Usage

```bash
jules /generate:tests --path="./OlympusActors-Cognition" --type="unit"
```
