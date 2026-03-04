@echo off
REM Unified Fleet Substrate Wasm-Native Tooling
REM Managed by: fleet-substrate
"%~dp0wazero-runner.exe" "%~dp0fleet-substrate.wasm" %*
