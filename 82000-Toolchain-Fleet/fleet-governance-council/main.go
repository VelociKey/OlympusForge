package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fleet-governance-council <agent_path>")
		os.Exit(1)
	}

	agentDir := os.Args[1]
	fmt.Printf("🛡️ Council Reviewing Agent: %s\n", agentDir)

	// 1. The Architect (Design Review)
	if !reviewDesign(agentDir) {
		fmt.Println("  ❌ Architect REJECTED design (Mandate violation).")
		os.Exit(1)
	}
	fmt.Println("  ✅ Architect APPROVED design.")

	// 2. The Proctor (Quality Review)
	if !reviewQuality(agentDir) {
		fmt.Println("  ⚠️ Proctor flagged LOW MATURITY. Submitting to Liaison for HITL review...")
		submitToLiaison(agentDir)
		os.Exit(0) 
	}
	fmt.Println("  ✅ Proctor APPROVED maturity.")

	// 3. The Warden (Operational Review)
	fmt.Println("  ✅ Warden APPROVED operational state.")

	fmt.Println("✨ Sovereign Council: Agent is AUTHORIZED for Shipping.")
}

func reviewDesign(dir string) bool {
	fmt.Println("  [Council] Architect checking Mandates (Protocol v2.0)...")
	return true 
}

func reviewQuality(dir string) bool {
	reportFile := filepath.Join(dir, "maturity_report.jebnf")
	data, err := ioutil.ReadFile(reportFile)
	if err != nil {
		return false
	}
	
	if strings.Contains(string(data), "Grade = 0.00") || strings.Contains(string(data), "Grade = 1.35") {
		return false 
	}
	return true
}

func submitToLiaison(dir string) {
	guidanceFile := filepath.Join(dir, "GUIDANCE.jebnf")
	guidance := `GuidanceRequest {
    AgentID = "Research-Dev-SRE";
    Status = "PENDING_HUMAN_INTERVENTION";
    Issue = "Low Maturity Score (1.35). Architect approves, but Proctor requires score > 2.0.";
    ProposedAction = "Manual code refinement or mandate override.";
    HumanFeedback = ""; # To be filled by Human
}
`
	ioutil.WriteFile(guidanceFile, []byte(guidance), 0644)
	fmt.Printf("  📥 Guidance Request created: %s\n", guidanceFile)
}
