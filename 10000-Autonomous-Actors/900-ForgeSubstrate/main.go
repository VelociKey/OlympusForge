package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Olympus Forge: Substrate Provisioner")
	fmt.Println("Forging the workstation-local GCP infrastructure...")

	composePath := "../../60000-Information-Storage/900-Emulators/podman-compose.yaml"

	cmd := exec.Command("podman-compose", "-f", composePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Starting Substrate via Podman Compose...")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start substrate: %v", err)
	}

	fmt.Println("\n--- Substrate Active ---")
	fmt.Println("Pub/Sub:      localhost:8085")
	fmt.Println("Datastore:    localhost:8090")
	fmt.Println("Bigtable:     localhost:8086")
	fmt.Println("Firestore:    localhost:8081")
	fmt.Println("Spanner:      localhost:9010 (gRPC), 9020 (HTTP)")
	fmt.Println("Storage:      localhost:4443")
	fmt.Println("Secret Manager (Vault): localhost:8200")
	fmt.Println("Vertex AI (Ollama): localhost:11434")
	fmt.Println("------------------------")
}
