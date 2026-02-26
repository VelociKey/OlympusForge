package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CycleData struct {
	ID        string
	Narrative string
	HumanTime time.Duration
	AiTime    time.Duration
}

func main() {
	now := time.Now()
	// If early morning, assume we want yesterday's work
	if now.Hour() < 4 {
		now = now.AddDate(0, 0, -1)
	}
	dateStr := now.Format("2006-01-02")

	fmt.Printf("Generating Fleet Daily Chronicle for %s...\n", dateStr)

	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v", err)
	}

	tracksDir := filepath.Join(root, "conductor", "tracks")
	data := collectDailyWork(tracksDir, now)

	var totalHuman, totalAi time.Duration
	var contextBuilder strings.Builder
	for _, d := range data {
		totalHuman += d.HumanTime
		totalAi += d.AiTime
		contextBuilder.WriteString(fmt.Sprintf("Track %s:\n%s\n\n", d.ID, d.Narrative))
	}

	aiSummary := callGemini(contextBuilder.String(), root)

	chronicle := formatChronicle(dateStr, aiSummary, totalHuman, totalAi)

	outputDir := filepath.Join(root, "PublicOutreach", "C0500-Agent-Intelligence-Outputs", "Daily-Summaries")
	os.MkdirAll(outputDir, 0755)

	outputPath := filepath.Join(outputDir, dateStr+".md")
	err = os.WriteFile(outputPath, []byte(chronicle), 0644)
	if err != nil {
		log.Fatalf("failed to write chronicle: %v", err)
	}

	fmt.Printf("Chronicle generated: %s\n", outputPath)
}

func collectDailyWork(tracksDir string, day time.Time) []CycleData {
	var data []CycleData
	
	filepath.WalkDir(tracksDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Base(path) != "plan.md" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Pick files modified today (or the target day)
		if info.ModTime().Year() != day.Year() || info.ModTime().YearDay() != day.YearDay() {
			return nil
		}

		content, _ := os.ReadFile(path)
		sContent := string(content)
		
		narrative := extractNarrative(sContent)
		if narrative == "" {
			return nil
		}

		trackID := filepath.Base(filepath.Dir(path))
		hTime, aTime := parseMetrics(filepath.Dir(path), trackID)

		data = append(data, CycleData{
			ID:        trackID,
			Narrative: narrative,
			HumanTime: hTime,
			AiTime:    aTime,
		})
		
		return nil
	})
	return data
}

func extractNarrative(content string) string {
	headers := []string{
		"## Semantic Narrative",
		"### Semantic Narrative",
		"## Narrative",
		"### Narrative",
	}

	for _, header := range headers {
		if strings.Contains(content, header) {
			parts := strings.Split(content, header)
			if len(parts) >= 2 {
				narrative := strings.TrimSpace(parts[1])
				if nextSection := strings.Index(narrative, "\n##"); nextSection != -1 {
					narrative = narrative[:nextSection]
				}
				if nextSection := strings.Index(narrative, "\n---"); nextSection != -1 {
					narrative = narrative[:nextSection]
				}
				return strings.TrimSpace(narrative)
			}
		}
	}
	return ""
}

func parseMetrics(dir string, id string) (human time.Duration, ai time.Duration) {
	metricsPath := filepath.Join(dir, fmt.Sprintf("Cycle-Metrics-%s.jebnf", id))
	content, err := os.ReadFile(metricsPath)
	if err != nil {
		return 0, 0
	}
	s := string(content)

	reHuman := regexp.MustCompile(`\(human "([^"]+)"\)`)
	reAi := regexp.MustCompile(`\(ai "([^"]+)"\)`)
	
	// Support for (start "...") and (end "...")
	reStart := regexp.MustCompile(`\(start "([^"]+)"\)`)
	reEnd := regexp.MustCompile(`\(end "([^"]+)"\)`)

	if m := reHuman.FindStringSubmatch(s); len(m) > 1 {
		human, _ = time.ParseDuration(m[1])
	}
	
	if m := reAi.FindStringSubmatch(s); len(m) > 1 {
		ai, _ = time.ParseDuration(m[1])
	} else {
		// Calculate AI time as sum of (end - start) if (ai "...") is missing
		starts := reStart.FindAllStringSubmatch(s, -1)
		ends := reEnd.FindAllStringSubmatch(s, -1)
		
		for i := 0; i < len(starts) && i < len(ends); i++ {
			startTime, errS := time.Parse(time.RFC3339, starts[i][1])
			endTime, errE := time.Parse(time.RFC3339, ends[i][1])
			if errS == nil && errE == nil {
				ai += endTime.Sub(startTime)
			}
		}
	}

	return human, ai
}

func callGemini(context string, root string) string {
	if context == "" {
		return "Today was a day of quiet preparation and steady momentum."
	}

	prompt := fmt.Sprintf(`You are Tracy Kidder, author of "The Soul of a New Machine". 
Based on the following technical activity from the Olympus Fleet today, write a compelling, 
prose-style summary of the day's struggle and progress. 
Focus on the transformation of the work and the emerging "soul" of the machine.

MANDATE: The "Actual AI-Assisted Human time" should be calculated as the sum of all (end - start) durations found in the metrics.

CONTEXT:
%s

Output only the prose narrative.`, context)

	// Use absolute path to gemini-cli.cmd
	geminiPath := filepath.Join(root, "OlympusForge", "90000-Enablement-Labs", "000-Tools", "000-bin", "gemini-cli.cmd")
	
	cmd := exec.Command(geminiPath, "--model", "gemini-1.5-flash", "ask", prompt)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "The fleet continued its steady evolution, weaving new patterns into the substrate of its being. (AI Summary unavailable: " + err.Error() + " - " + stderr.String() + ")"
	}

	return strings.TrimSpace(out.String())
}

func formatChronicle(date string, narrative string, human time.Duration, ai time.Duration) string {
	var sb strings.Builder
	sb.WriteString("# All in a Day's Work: " + date + "\n")
	sb.WriteString("## *A Chronicle of the Sovereign Fleet*\n\n")
	sb.WriteString("> \"The machine was still a collection of parts, but it was beginning to have a soul.\" - In the spirit of Tracy Kidder\n\n")
	
	sb.WriteString("### The Narrative\n")
	sb.WriteString(narrative + "\n\n")

	sb.WriteString("### Labor & Metrics\n")
	sb.WriteString(fmt.Sprintf("- **Estimated Human Labor**: %s\n", formatHHMM(human)))
	sb.WriteString(fmt.Sprintf("- **Actual AI-Assisted Time**: %s\n", formatHHMM(ai)))
	
	efficiency := 0.0
	if ai.Minutes() > 0 {
		efficiency = human.Minutes() / ai.Minutes()
	}
	sb.WriteString(fmt.Sprintf("- **Sovereign Efficiency Multiplier**: %.1fx\n\n", efficiency))

	sb.WriteString("---\n")
	sb.WriteString("**Generated by fleet-daily-chronicle.exe**\n")
	sb.WriteString("**Fleet Timestamp**: " + time.Now().Format(time.RFC1123) + "  \n")
	sb.WriteString("**Location**: `PublicOutreach/C0500-Agent-Intelligence-Outputs/Daily-Summaries/`\n")
	
	return sb.String()
}

func formatHHMM(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}
