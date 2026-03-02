package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	econotel "olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/P0000-pkg/000-econotel"
	fleet "olympus.fleet/00SDLC/OlympusForge/00000-Identity-Foundations/P0000-pkg/000-fleet"

	reasoningv1 "olympus.fleet/00SDLC/OlympusGrammar/gen/v1/reasoning"
	reasoningv1connect "olympus.fleet/00SDLC/OlympusGrammar/gen/v1/reasoning/reasoningv1connect"

	"net/http"

	"io"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	cwd, _ := os.Getwd()
	var root string

	// Traverse up from CWD until we find the folder containing "Olympus2"
	// This ensures we always find the absolute root of aAntigravitySpace
	curr := cwd
	for {
		if _, err := os.Stat(filepath.Join(curr, "Olympus2")); err == nil {
			root = curr
			break
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			root = cwd // Fallback
			break
		}
		curr = parent
	}

	switch command {
	case "init":
		runInit(root)
	case "tidy":
		runTidy(root)
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: sovereign create <moduleName>")
			os.Exit(1)
		}
		runCreate(root, os.Args[2])
	case "assess":
		if len(os.Args) < 3 {
			runAssess(root, "all")
		} else {
			runAssess(root, os.Args[2])
		}
	case "check":
		runCheck(root)
	case "start":
		runStart(root)
	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Usage: sovereign build <moduleName>")
			os.Exit(1)
		}
		runBuild(root, os.Args[2])
	case "docs":
		runDocs(root)
	case "watch":
		runWatch()
	case "sync":
		runSync(root)
	case "migrate":
		if len(os.Args) < 4 {
			fmt.Println("Usage: sovereign migrate <oldPath> <newPath>")
			os.Exit(1)
		}
		runMigrate(root, os.Args[2], os.Args[3])
	case "daemon":
		runDaemon(root)
	case "ask":
		if len(os.Args) < 3 {
			fmt.Println("Usage: sovereign ask <question>")
			os.Exit(1)
		}
		runAsk(root, os.Args[2])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Sovereign Fleet CLI")
	fmt.Println("Usage: sovereign <command> [args]")
	fmt.Println("\nCommands:")
	fmt.Println("  init    Synchronize all go.mod and go.work files")
	fmt.Println("  tidy    Run 'go mod tidy' across all modules")
	fmt.Println("  create  Scaffold a new Go module")
	fmt.Println("  assess  Assess maturity of workspace(s) via Athena (Dagger)")
	fmt.Println("  check   Verify fleet integrity and structural consistency")
	fmt.Println("  start   Start the Sovereign Workstation Cloud (Infrastructure + Bridges)")
	fmt.Println("  build   Build a module natively leveraging OlympusForge")
	fmt.Println("  docs    Aggregate markdown documentation for the active workspace")
	fmt.Println("  watch   Open the interactive Mesh Watcher dashboard")
	fmt.Println("  sync    Ping Sovereign context to the shared Conductor LPSV session log")
	fmt.Println("  migrate Update import paths and replacements across the fleet")
	fmt.Println("  daemon  Start the Sovereign background watcher daemon")
	fmt.Println("  ask     Ask George for assistance (context-aware)")
	fmt.Println("  help    Show this message")
}

func runStart(root string) {
	fmt.Println("🚀 Starting Sovereign Cloud...")
	cmd := exec.Command("go", "run", "orchestrate.go")
	cmd.Dir = filepath.Join(root, "OlympusInfrastructure")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error starting cloud: %v\n", err)
		os.Exit(1)
	}
}

func runCheck(root string) {
	if err := fleet.VerifyFleet(root); err != nil {
		os.Exit(1)
	}
}

func runCreate(root, moduleName string) {
	if err := fleet.ScaffoldModule(root, moduleName); err != nil {
		fmt.Printf("Error creating module: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Module scaffolded. Initializing fleet...")
	runInit(root)
	fmt.Println("🧹 Module ready. running tidy...")
	runTidy(root)
}

func runMigrate(root, oldPath, newPath string) {
	fmt.Printf("🚀 Migrating imports from %s to %s...\n", oldPath, newPath)
	if err := fleet.MigrateImports(root, oldPath, newPath); err != nil {
		fmt.Printf("Error during migration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Migration complete. Tidying dependencies...")
	runTidy(root)
}

func runInit(root string) {
	fmt.Println("🚀 Initializing Sovereign Fleet...")

	modules, err := fleet.FindGoModules(root)
	if err != nil {
		fmt.Printf("Error scanning modules: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d modules. Syncing go.mod files...\n", len(modules))
	for _, mod := range modules {
		if err := fleet.SyncGoMod(mod, modules, root); err != nil {
			fmt.Printf("Error syncing %s: %v\n", mod.Name, err)
		}
	}

	fmt.Println("Syncing go.work...")
	if err := fleet.SyncGoWork(root, modules); err != nil {
		fmt.Printf("Error syncing go.work: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Fleet initialization complete.")
}

func runTidy(root string) {
	fmt.Println("🧹 Tidying Fleet Dependencies...")

	modules, err := fleet.FindGoModules(root)
	if err != nil {
		fmt.Printf("Error scanning modules: %v\n", err)
		os.Exit(1)
	}

	for _, mod := range modules {
		fmt.Printf("Tidying %s...\n", mod.Name)
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = mod.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: 'go mod tidy' failed in %s: %v\n", mod.Name, err)
		}
	}

	fmt.Println("Synchronizing workspace...")
	cmd := exec.Command("go", "work", "sync")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: 'go work sync' failed: %v\n", err)
	}

	fmt.Println("✅ Fleet tidying complete.")
}

func runBuild(root, moduleName string) {
	scratchDir := filepath.Join(root, "Olympus2", "C0990-Ephemeral-Scratch")
	os.MkdirAll(scratchDir, 0755)
	lastErrFile := filepath.Join(scratchDir, "last_error.log")

	fmt.Printf("🚀 Building %s natively via OlympusForge...\n", moduleName)
	cmd := exec.Command("go", "run", ".", "-target", "native", "-workspace", moduleName)
	cmd.Dir = filepath.Join(root, "OlympusForge", "90000-Enablement-Labs", "900-Forge")

	f, _ := os.Create(lastErrFile)
	defer f.Close()

	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, f)

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error building module: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Build pipeline complete.")
}

func runAssess(root, workspace string) {
	fmt.Printf("🛡️  Assessing maturity of %s via Athena...\n", workspace)
	cmd := exec.Command("go", "run", ".", "-assess", "-workspace", workspace)
	cmd.Dir = filepath.Join(root, "OlympusForge", "90000-Enablement-Labs", "900-Forge")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error during assessment: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Assessment complete.")
}

func runDocs(root string) {
	fmt.Println("📚 Aggregating Sovereign Workspace Documentation...")
	modules, err := fleet.FindGoModules(root)
	if err != nil {
		fmt.Printf("Error scanning modules: %v\n", err)
		os.Exit(1)
	}

	for _, mod := range modules {
		docFile := filepath.Join(mod.Path, "SOVEREIGN_DOCS.md")

		var content []byte
		filepath.Walk(mod.Path, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && filepath.Ext(path) == ".md" && filepath.Base(path) != "SOVEREIGN_DOCS.md" {
				data, readErr := os.ReadFile(path)
				if readErr == nil {
					content = append(content, []byte(fmt.Sprintf("\n# Source: %s\n\n", filepath.Base(path)))...)
					content = append(content, data...)
				}
			}
			return nil
		})

		if len(content) > 0 {
			os.WriteFile(docFile, content, 0644)
			fmt.Printf("Generated %s\n", docFile)
		}
	}
	fmt.Println("✅ Documentation aggregation complete.")
}

func runWatch() {
	ports := []string{"8092", "8091", "8096", "8098", "8095", "8093", "8094", "8097", "8099", "8090"}

	for {
		// Clear screen
		fmt.Print("\033[H\033[2J")
		fmt.Println("📡 Sovereign Mesh Watcher Dashboard")
		fmt.Println("=====================================")
		fmt.Printf("Time: %s\n\n", time.Now().Format(time.RFC1123))

		active := 0
		for _, port := range ports {
			conn, err := net.DialTimeout("tcp", "localhost:"+port, 500*time.Millisecond)
			if err == nil {
				fmt.Printf(" [OK] Port %s: 🟩 ONLINE\n", port)
				active++
				conn.Close()
			} else {
				fmt.Printf(" [XX] Port %s: 🟥 OFFLINE\n", port)
			}
		}

		fmt.Println("-------------------------------------")
		fmt.Printf("Total Online: %d/%d\n", active, len(ports))
		fmt.Println("Press Ctrl+C to exit.")

		time.Sleep(2 * time.Second)
	}
}

func runSync(root string) {
	fmt.Println("📡 Pinging Sovereign Context to Conductor LPSV Session Log...")
	logFile := filepath.Join(root, "conductor", "session.lpsv")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening session log: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	tp, err := econotel.InitTracer("conductor-sidecar", f)
	if err != nil {
		fmt.Printf("Error initializing econotel: %v\n", err)
		os.Exit(1)
	}
	defer tp.Shutdown(context.Background())

	tr := tp.Tracer("sovereign-sync")
	ctx, span := tr.Start(context.Background(), "SyncContext")
	defer span.End()

	econotel.RecordEconomicEvent(ctx, "session.sync", 1.0, "ping", 0.0)
	span.SetAttributes(attribute.String("fleet.workspace", root))
	span.SetAttributes(attribute.String("message", "Sovereign CLI synchronized workspace attributes"))

	fmt.Printf("✅ Synced Context Ping to %s\n", logFile)
}

func runDaemon(root string) {
	fmt.Println("👁️  Starting Sovereign Daemon...")
	// In a real implementation, this would use fsnotify to watch for changes
	// and update George's knowledge base or trigger builds.
	for {
		fmt.Printf("[%s] Sovereign Daemon active. Watching for fleet events...\n", time.Now().Format("15:04:05"))
		time.Sleep(10 * time.Second)
	}
}

func runAsk(root string, question string) {
	fmt.Printf("🧠 Consulting George: %s\n", question)

	// Check for active failures in scratch space
	scratchDir := filepath.Join(root, "Olympus2", "C0990-Ephemeral-Scratch")
	lastErrFile := filepath.Join(scratchDir, "last_error.log")
	var contextStr string
	if data, err := os.ReadFile(lastErrFile); err == nil {
		fmt.Println("📎 Detected active failure in scratch, adding to context...")
		contextStr = fmt.Sprintf("\n\nRECENT_FAILURE_LOG:\n%s", string(data))
	}

	client := reasoningv1connect.NewAutonomousInferenceServiceClient(http.DefaultClient, "http://localhost:8080")

	// 1. Start or resume session (mocking user_id for now)
	ctx := context.Background()
	startReq := connect.NewRequest(&reasoningv1.StartSessionRequest{
		UserId: "sovereign_cli_user",
	})
	startReq.Header().Set("Authorization", "Bearer sovereign_cli_user-token")

	startResp, err := client.StartSession(ctx, startReq)
	if err != nil {
		fmt.Printf("❌ Failed to connect to George: %v\n", err)
		return
	}
	sessionID := startResp.Msg.SessionId

	// 2. Send message
	req := connect.NewRequest(&reasoningv1.SendMessageRequest{
		SessionId: sessionID,
		Text:      question + contextStr,
	})
	req.Header().Set("Authorization", "Bearer sovereign_cli_user-token")

	stream, err := client.SendMessage(ctx, req)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Print("🤖 George: ")
	for stream.Receive() {
		event := stream.Msg()
		if event.Type == reasoningv1.AgentEvent_TOKEN {
			fmt.Print(event.Payload)
		}
	}
	fmt.Println()

	if err := stream.Err(); err != nil {
		fmt.Printf("\n❌ Stream error: %v\n", err)
	}
}
