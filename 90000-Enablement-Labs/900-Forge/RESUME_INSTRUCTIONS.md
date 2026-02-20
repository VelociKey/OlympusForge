# AIHub "Master Forge" Resumption Instructions

**Status:** Halted due to network instability during `debian:trixie` provisioning.
**Last Action:** Downloading `ollama-linux-amd64.tar.zst` inside the Dagger container.

## How to Resume

1.  **Verify Network:** Ensure stable internet connectivity to `deb.debian.org`, `github.com`, and `ollama.com`.
2.  **Verify Resources:** Ensure Podman machine is running with 8GB RAM:
    ```bash
    podman machine info
    ```
3.  **Resume Dagger Pipeline:**
    Run the Go-native Dagger script. The build cache should pick up *after* the base image and Go installation, re-attempting only the failed steps (Ollama download).
    ```bash
    cd c:\AIHubDevelopment\aihub-forge
    go run main.go
    ```

## Current Configuration State

*   **Base Image:** `debian:trixie-slim` (Debian 13 Testing/Stable Candidate)
*   **Foundation:** Go 1.25 (via 1.24 installer logic), Git, `zstd` (added for Ollama extraction).
*   **Target:** Local Docker Engine (`npipe:////./pipe/docker_engine`).
*   **Publication:** Configured to push to `AIHUB_GAR_DEST` if set; otherwise verifying locally.

## Troubleshooting

If the network hangs persist:
*   **Clean Cache:** `dagger call clean` (if applicable) or simply restart the Podman machine:
    ```bash
    podman machine stop
    podman machine start
    ```
