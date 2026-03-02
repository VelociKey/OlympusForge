@echo off
setlocal
set GOCACHE=%~dp0..\..\..\..\00SDLC\Olympus2\C0990-Ephemeral-Scratch\go-cache
"%~dp0..\..\..\81000-Toolchain-External\go\bin\go.exe" %*
endlocal
