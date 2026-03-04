package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Sovereign Rental Agent Provisioning...")
	// Standard provisioning: Check license, initialize local storage, etc.
	if _, err := os.Stat("AgentLease.jebnf"); os.IsNotExist(err) {
		fmt.Println("âŒ ERROR: Missing AgentLease.jebnf. Agent is non-sovereign.")
		os.Exit(1)
	}
	fmt.Println("âœ… Provisioning Complete.")
}
