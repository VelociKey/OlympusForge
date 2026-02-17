@echo off
if not exist "%~dp0..\infrastructure\trivy-cache" mkdir "%~dp0..\infrastructure\trivy-cache"
"%~dp0..\infrastructure\trivy.exe" --cache-dir "%~dp0..\infrastructure\trivy-cache" %*
