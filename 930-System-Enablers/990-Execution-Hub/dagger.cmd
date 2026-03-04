@echo off
set "DAGGER_CACHE_DIR=C:\aAntigravitySpace\00SDLC\Olympus2\C0990-Ephemeral-Scratch\DAGGER\cache"
set "DAGGER_CONFIG_HOME=C:\aAntigravitySpace\00SDLC\Olympus2\C0990-Ephemeral-Scratch\DAGGER\config"
set "DAGGER_ENGINE_IMAGE=registry.dagger.io/engine:v0.20.0"
if not exist "C:\aAntigravitySpace\00SDLC\Olympus2\C0990-Ephemeral-Scratch\DAGGER" mkdir "C:\aAntigravitySpace\00SDLC\Olympus2\C0990-Ephemeral-Scratch\DAGGER"
"C:\aAntigravitySpace\00SDLC\OlympusForge\90000-Enablement-Labs\000-Tools\000-external\dagger\dagger.exe" %*
