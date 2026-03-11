package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	slog.Info("📜 Initializing Fleet Daily Chronicle (Metrics Aggregator)")

	tracksDir := "conductor/tracks"
	var humanLaborTotal, aiLaborTotal time.Duration
	var humanInitiatedCount, aiInitiatedCount int
	tracksProcessed := 0

	err := filepath.Walk(tracksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !info.IsDir() && strings.HasPrefix(info.Name(), "Cycle-Metrics-CYC-") && strings.HasSuffix(info.Name(), ".jebnf") {
			h, a, initiator := parseMetrics(path)
			humanLaborTotal += h
			aiLaborTotal += a
			if initiator == "AI" {
				aiInitiatedCount++
			} else {
				humanInitiatedCount++
			}
			tracksProcessed++
		}
		return nil
	})

	if err != nil {
		slog.Error("Failed to walk tracks directory", "error", err)
		os.Exit(1)
	}

	report := formatReport(humanLaborTotal, aiLaborTotal, tracksProcessed, humanInitiatedCount, aiInitiatedCount)

	// Output to both a daily file and a global summary if needed
	outputDir := "00SDLC/Olympus2"
	outputPath := filepath.Join(outputDir, fmt.Sprintf("ALL_IN_A_DAYS_WORK_%s.md", time.Now().Format("20060102")))

	if err := os.WriteFile(outputPath, []byte(report), 0644); err != nil {
		slog.Error("Failed to write report", "error", err)
		os.Exit(1)
	}

	slog.Info("✅ Daily Chronicle Finalized", "path", outputPath, "tracks", tracksProcessed)
}

func parseMetrics(path string) (time.Duration, time.Duration, string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, "Human"
	}

	humanRegex := regexp.MustCompile(`TotalHumanEstimate\s+"([^"]+)"`)
	actualRegex := regexp.MustCompile(`TotalActualWork\s+"([^"]+)"`)
	initiatorRegex := regexp.MustCompile(`Initiator\s+=\s+"([^"]+)"`)

	humanMatch := humanRegex.FindStringSubmatch(string(content))
	actualMatch := actualRegex.FindStringSubmatch(string(content))
	initiatorMatch := initiatorRegex.FindStringSubmatch(string(content))

	var human, actual time.Duration
	if len(humanMatch) > 1 { human = parseDuration(humanMatch[1]) }
	if len(actualMatch) > 1 { actual = parseDuration(actualMatch[1]) }

	initiator := "Human" // Default
	if len(initiatorMatch) > 1 { initiator = initiatorMatch[1] }

	return human, actual, initiator
}

func parseDuration(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" { return 0 }
	if val, err := strconv.Atoi(s); err == nil { return time.Duration(val) * time.Minute }
	total := time.Duration(0)
	parts := strings.Fields(s)
	for _, part := range parts {
		d, err := time.ParseDuration(part)
		if err == nil { total += d } else {
			numStr, unit := "", ""
			for i, r := range part {
				if r >= '0' && r <= '9' { numStr += string(r) } else { unit = part[i:]; break }
			}
			num, _ := strconv.Atoi(numStr)
			switch unit {
			case "h": total += time.Duration(num) * time.Hour
			case "m": total += time.Duration(num) * time.Minute
			case "s": total += time.Duration(num) * time.Second
			}
		}
	}
	return total
}

func formatReport(human, ai time.Duration, tracks, hCount, aCount int) string {
	efficiency := 0.0
	if ai > 0 { efficiency = float64(human) / float64(ai) }
	dateStr := time.Now().Format("January 02, 2006")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# All In A Day's Work: Global Fleet\n"))
	sb.WriteString(fmt.Sprintf("**Date**: %s\n\n", dateStr))
	sb.WriteString(fmt.Sprintf("## 📊 Efficiency Multiplier\n"))
	sb.WriteString(fmt.Sprintf("- **Total Tracks**: %d (Human: %d, AI: %d)\n", tracks, hCount, aCount))
	sb.WriteString(fmt.Sprintf("- **Estimated Human Labor**: %s\n", formatDuration(human)))
	sb.WriteString(fmt.Sprintf("- **AI Actual Time**: %s\n", formatDuration(ai)))
	sb.WriteString(fmt.Sprintf("- **Overall Efficiency**: %.2fx\n\n", efficiency))

	sb.WriteString("## 📜 Semantic Narrative\n")
	sb.WriteString("The fleet has successfully integrated the **Sovereign Sentry** autonomous self-improvement loop. By differentiating between Human and AI-initiated cycles, we gain deep visibility into the fleet's evolutionary velocity. The metrics reflect a significant acceleration in delivery speed through agentic orchestration and the liquidation of legacy shell dependencies.\n")

	return sb.String()
}
func formatDuration(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	
	result := ""
	if h > 0 {
		result += fmt.Sprintf("%dh ", h)
	}
	if m > 0 || h > 0 {
		result += fmt.Sprintf("%dm ", m)
	}
	result += fmt.Sprintf("%ds", s)
	return strings.TrimSpace(result)
}
