package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Prerequisite struct {
	Name    string
	Command string
	Args    []string
	ID      string
}

type ToolSpec struct {
	Version         string
	WinURL          string
	MacArmURL       string
	MacIntelURL     string
	LinuxURL        string
	TargetSubFolder string // Folder inside 81000
	Purify          bool
	PostInstall     func(string) error // toolchainDir as arg
}

var prereqs = []Prerequisite{
	{Name: "Go Runtime", Command: "go", Args: []string{"version"}, ID: "GoLang.Go"},
	{Name: "Podman CLI", Command: "podman", Args: []string{"--version"}, ID: "RedHat.Podman-Desktop"},
	{Name: "Git CLI", Command: "git", Args: []string{"--version"}, ID: "Git.Git"},
	{Name: "GitHub CLI", Command: "gh", Args: []string{"--version"}, ID: "GitHub.cli"},
	{Name: "JQ", Command: "jq", Args: []string{"--version"}, ID: "jqlang.jq"},
	{Name: "DuckDB", Command: "duckdb", Args: []string{"--version"}, ID: "DuckDB.DuckDB"},
	{Name: "Trivy", Command: "trivy", Args: []string{"--version"}, ID: "AquaSecurity.Trivy"},
	{Name: "OpenTofu", Command: "tofu", Args: []string{"--version"}, ID: "OpenTofu.OpenTofu"},
	{Name: "Java JDK 25", Command: "java", Args: []string{"-version"}, ID: "Oracle.JDK.25"},
	{Name: "Firebase CLI", Command: "firebase", Args: []string{"--version"}, ID: "firebase-tools"},
}

var toolSpecs = map[string]ToolSpec{
	"GoLang.Go": {
		Version:         "1.26.0",
		WinURL:          "https://go.dev/dl/go1.26.0.windows-amd64.zip",
		MacArmURL:       "https://go.dev/dl/go1.26.0.darwin-arm64.tar.gz",
		MacIntelURL:     "https://go.dev/dl/go1.26.0.darwin-amd64.tar.gz",
		LinuxURL:        "https://go.dev/dl/go1.26.0.linux-amd64.tar.gz",
		TargetSubFolder: "go",
		Purify:          false, // Go is too sensitive for broad purification
		PostInstall: func(toolchainDir string) error {
			goDir := filepath.Join(toolchainDir, "go")

			// Flatten nested go/go folder if gemaid fetch extracted it that way
			nestedGo := filepath.Join(goDir, "go")
			if _, err := os.Stat(nestedGo); err == nil {
				tmp := goDir + "_tmp"
				os.Rename(nestedGo, tmp)
				os.RemoveAll(goDir)
				os.Rename(tmp, goDir)
			}

			// Resolve absolute paths for scratch directories
			wd, _ := os.Getwd()
			root, _ := filepath.Abs(wd)
			for {
				if _, err1 := os.Stat(filepath.Join(root, "00SDLC")); err1 == nil {
					if _, err2 := os.Stat(filepath.Join(root, "10GNDT")); err2 == nil {
						break // Found true fleet root
					}
				}
				parent := filepath.Dir(root)
				if parent == root {
					break
				}
				root = parent
			}

			scratch := filepath.Join(root, "00SDLC", "Olympus2", "C0990-Ephemeral-Scratch")

			goCache := filepath.Join(scratch, "go-cache")
			goPath := filepath.Join(scratch, "go-path")
			os.MkdirAll(goCache, 0755)
			os.MkdirAll(goPath, 0755)

			goEnvPath := filepath.Join(goDir, "go.env")
			envContent := fmt.Sprintf("GOPROXY=https://proxy.golang.org,direct\nGOSUMDB=sum.golang.org\nGOTOOLCHAIN=local\nGOCACHE=%s\nGOPATH=%s\n", goCache, goPath)
			if err := os.WriteFile(goEnvPath, []byte(envContent), 0644); err != nil {
				return err
			}

			// Update VS Code settings
			vscodeSettings := filepath.Join(root, ".vscode", "settings.json")
			if _, err := os.Stat(vscodeSettings); err == nil {
				data, _ := os.ReadFile(vscodeSettings)
				s := string(data)
				if !strings.Contains(s, `"go.gopath"`) {
					formattedGoPath := strings.ReplaceAll(goPath, "\\", "\\\\")
					replacement := fmt.Sprintf("\"go.gopath\": \"%s\",\n    \"go.goroot\":", formattedGoPath)
					s = strings.Replace(s, `"go.goroot":`, replacement, 1)
					os.WriteFile(vscodeSettings, []byte(s), 0644)
				}
			}

			return nil
		},
	},
	"RedHat.Podman-Desktop": {
		Version:         "5.8.0",
		WinURL:          "https://github.com/containers/podman/releases/download/v5.8.0/podman-remote-release-windows_amd64.zip",
		MacArmURL:       "https://github.com/containers/podman/releases/download/v5.8.0/podman-remote-release-darwin_arm64.zip",
		MacIntelURL:     "https://github.com/containers/podman/releases/download/v5.8.0/podman-remote-release-darwin_amd64.zip",
		TargetSubFolder: "podman",
		PostInstall: func(toolchainDir string) error {
			dirs, _ := os.ReadDir(toolchainDir)
			var extracted string
			for _, d := range dirs {
				if d.IsDir() && strings.HasPrefix(d.Name(), "podman-") && d.Name() != "podman" {
					extracted = filepath.Join(toolchainDir, d.Name())
					break
				}
			}
			if extracted == "" {
				return nil
			}
			binDir := filepath.Join(extracted, "usr", "bin")
			if runtime.GOOS != "windows" {
				binDir = filepath.Join(extracted, "bin")
			}
			target := filepath.Join(toolchainDir, "podman")
			os.MkdirAll(target, 0755)
			files, _ := os.ReadDir(binDir)
			for _, f := range files {
				os.Rename(filepath.Join(binDir, f.Name()), filepath.Join(target, f.Name()))
			}
			os.RemoveAll(extracted)
			return nil
		},
	},
	"Git.Git": {
		Version:         "2.53.0.windows.1",
		WinURL:          "https://github.com/git-for-windows/git/releases/download/v2.53.0.windows.1/MinGit-2.53.0-64-bit.zip",
		MacArmURL:       "https://github.com/timcharper/git-osx-installer/releases/download/v2.39.2/git-2.39.2-arm64-bit.zip",
		TargetSubFolder: "git",
	},
	"GitHub.cli": {
		Version:         "2.87.3",
		WinURL:          "https://github.com/cli/cli/releases/download/v2.87.3/gh_2.87.3_windows_amd64.zip",
		MacArmURL:       "https://github.com/cli/cli/releases/download/v2.87.3/gh_2.87.3_macOS_arm64.zip",
		MacIntelURL:     "https://github.com/cli/cli/releases/download/v2.87.3/gh_2.87.3_macOS_amd64.zip",
		TargetSubFolder: "gh",
		PostInstall: func(toolchainDir string) error {
			target := filepath.Join(toolchainDir, "gh")
			os.MkdirAll(target, 0755)
			filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() && (d.Name() == "gh" || d.Name() == "gh.exe") {
					os.Rename(path, filepath.Join(target, d.Name()))
				}
				return nil
			})
			return nil
		},
	},
	"jqlang.jq": {
		Version:         "1.8.1",
		WinURL:          "https://github.com/jqlang/jq/releases/download/jq-1.8.1/jq-windows-amd64.exe",
		MacArmURL:       "https://github.com/jqlang/jq/releases/download/jq-1.8.1/jq-macos-arm64",
		MacIntelURL:     "https://github.com/jqlang/jq/releases/download/jq-1.8.1/jq-macos-amd64",
		TargetSubFolder: "jq",
		PostInstall: func(toolchainDir string) error {
			target := filepath.Join(toolchainDir, "jq")
			files, _ := os.ReadDir(target)
			for _, f := range files {
				if strings.Contains(f.Name(), "jq") && !f.IsDir() {
					newName := "jq"
					if runtime.GOOS == "windows" {
						newName += ".exe"
					}
					os.Rename(filepath.Join(target, f.Name()), filepath.Join(target, newName))
					os.Chmod(filepath.Join(target, newName), 0755)
				}
			}
			return nil
		},
	},
	"DuckDB.DuckDB": {
		Version:         "1.4.4",
		WinURL:          "https://github.com/duckdb/duckdb/releases/download/v1.4.4/duckdb_cli-windows-amd64.zip",
		MacArmURL:       "https://github.com/duckdb/duckdb/releases/download/v1.4.4/duckdb_cli-osx-arm64.zip",
		TargetSubFolder: "duckdb",
	},
	"AquaSecurity.Trivy": {
		Version:         "0.69.3",
		WinURL:          "https://github.com/aquasecurity/trivy/releases/download/v0.69.3/trivy_0.69.3_windows-64bit.zip",
		MacArmURL:       "https://github.com/aquasecurity/trivy/releases/download/v0.69.3/trivy_0.69.3_macOS-ARM64.tar.gz",
		TargetSubFolder: "trivy",
	},
	"OpenTofu.OpenTofu": {
		Version:         "1.11.5",
		WinURL:          "https://github.com/opentofu/opentofu/releases/download/v1.11.5/tofu_1.11.5_windows_amd64.zip",
		MacArmURL:       "https://github.com/opentofu/opentofu/releases/download/v1.11.5/tofu_1.11.5_darwin_arm64.zip",
		TargetSubFolder: "tofu",
	},
	"Oracle.JDK.25": {
		Version:         "25.0.2+10",
		WinURL:          "https://github.com/adoptium/temurin25-binaries/releases/download/jdk-25.0.2%2B10/OpenJDK25U-jdk_x64_windows_hotspot_25.0.2_10.zip",
		MacArmURL:       "https://github.com/adoptium/temurin25-binaries/releases/download/jdk-25.0.2%2B10/OpenJDK25U-jdk_aarch64_mac_hotspot_25.0.2_10.tar.gz",
		TargetSubFolder: "java",
		PostInstall: func(toolchainDir string) error {
			target := filepath.Join(toolchainDir, "java")
			items, _ := os.ReadDir(target)
			for _, item := range items {
				if item.IsDir() {
					inner := filepath.Join(target, item.Name())
					contents, _ := os.ReadDir(inner)
					for _, c := range contents {
						os.Rename(filepath.Join(inner, c.Name()), filepath.Join(target, c.Name()))
					}
					os.RemoveAll(inner)
					break
				}
			}
			return nil
		},
	},
	"firebase-tools": {
		Version:         "latest",
		WinURL:          "https://firebase.tools/bin/win/latest",
		MacArmURL:       "https://firebase.tools/bin/macos/latest",
		TargetSubFolder: "firebase",
		PostInstall: func(toolchainDir string) error {
			target := filepath.Join(toolchainDir, "firebase")
			files, _ := os.ReadDir(target)
			for _, f := range files {
				if !f.IsDir() {
					newName := "firebase"
					if runtime.GOOS == "windows" {
						newName += ".exe"
					}
					os.Rename(filepath.Join(target, f.Name()), filepath.Join(target, newName))
					os.Chmod(filepath.Join(target, newName), 0755)
				}
			}
			return nil
		},
	},
}

func main() {
	install := flag.Bool("install", false, "Install missing prerequisites")
	update := flag.Bool("update", false, "Update existing prerequisites to latest version")
	scorch := flag.Bool("scorch", false, "Report legacy binaries found outside the Forge")
	flag.Parse()

	fmt.Println("🚀 Olympus Fleet Bootstrap: Sovereign Alignment")

	for _, p := range prereqs {
		fmt.Printf("🔍 Checking: %s... ", p.Name)
		if isInstalled(p.Command, p.Args...) {
			fmt.Println("INSTALLED")
			if *update {
				installSovereignTool(p.ID)
			}
		} else {
			fmt.Println("MISSING")
			if *install {
				installSovereignTool(p.ID)
			}
		}
	}

	if *install {
		setupPodmanMachine()
	}

	if *scorch {
		runBinaryNormalizationReport()
	}

	fmt.Println("\n✅ Bootstrap sequence complete.")
}

func isInstalled(name string, args ...string) bool {
	p := resolveSovereignTool(name)
	if p == name {
		return false
	}
	cmd := exec.Command(p, args...)
	err := cmd.Run()
	return err == nil
}

func resolveSovereignTool(name string) string {
	wd, _ := os.Getwd()
	toolchainDir := filepath.Join(wd, "81000-Toolchain-External")
	if filepath.Base(wd) == "fleet-bootstrap" || filepath.Base(wd) == "bin" {
		toolchainDir = filepath.Clean(filepath.Join(wd, "..", "..", "81000-Toolchain-External"))
	}

	exeName := name
	if runtime.GOOS == "windows" && !strings.HasSuffix(exeName, ".exe") {
		exeName += ".exe"
	}

	var candidatePaths []string
	switch name {
	case "go":
		candidatePaths = []string{filepath.Join(toolchainDir, "go", "bin", exeName)}
	case "podman":
		candidatePaths = []string{filepath.Join(toolchainDir, "podman", exeName)}
	case "git":
		if runtime.GOOS == "windows" {
			candidatePaths = []string{filepath.Join(toolchainDir, "git", "cmd", "git.exe")}
		} else {
			candidatePaths = []string{filepath.Join(toolchainDir, "git", "bin", "git")}
		}
	case "gh":
		candidatePaths = []string{filepath.Join(toolchainDir, "gh", exeName)}
	case "jq":
		candidatePaths = []string{filepath.Join(toolchainDir, "jq", exeName)}
	case "duckdb":
		candidatePaths = []string{filepath.Join(toolchainDir, "duckdb", exeName)}
	case "trivy":
		candidatePaths = []string{filepath.Join(toolchainDir, "trivy", exeName)}
	case "tofu":
		candidatePaths = []string{filepath.Join(toolchainDir, "tofu", exeName)}
	case "java":
		candidatePaths = []string{filepath.Join(toolchainDir, "java", "bin", exeName)}
	case "firebase":
		candidatePaths = []string{filepath.Join(toolchainDir, "firebase", exeName)}
	}

	for _, p := range candidatePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return name
}

func installSovereignTool(id string) {
	spec, ok := toolSpecs[id]
	if !ok {
		fmt.Printf("⚠️  No Sovereign spec for %s, skipping.\n", id)
		return
	}

	fmt.Printf("🛠️  Installing Sovereign %s (%s)... ", id, spec.Version)

	wd, _ := os.Getwd()
	toolchainDir := filepath.Join(wd, "81000-Toolchain-External")
	if filepath.Base(wd) == "fleet-bootstrap" || filepath.Base(wd) == "bin" {
		toolchainDir = filepath.Clean(filepath.Join(wd, "..", "..", "81000-Toolchain-External"))
	}

	gemaidPath := filepath.Join(wd, "82000-Toolchain-Fleet", "bin", "gemaid.exe")
	if filepath.Base(wd) == "fleet-bootstrap" {
		gemaidPath = filepath.Join(wd, "..", "bin", "gemaid.exe")
	}
	if runtime.GOOS != "windows" {
		gemaidPath = strings.TrimSuffix(gemaidPath, ".exe")
	}

	url := spec.WinURL
	if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" && spec.MacArmURL != "" {
			url = spec.MacArmURL
		} else {
			url = spec.MacIntelURL
		}
	} else if runtime.GOOS == "linux" && spec.LinuxURL != "" {
		url = spec.LinuxURL
	}

	if url == "" {
		fmt.Println("FAILED (No URL for this platform)")
		return
	}

	targetDir := filepath.Join(toolchainDir, spec.TargetSubFolder)

	// If it's Go, we want a clean slate to avoid merging old/new versions, but handle safely
	if id == "GoLang.Go" {
		if _, err := os.Stat(targetDir); err == nil {
			if err := os.Rename(targetDir, targetDir+".old"); err != nil {
				fmt.Printf("ERROR: Failed to rotate old toolchain: %v\n", err)
				return
			}
			defer os.RemoveAll(targetDir + ".old")
		}
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("ERROR: Failed to mkdir %s: %v\n", targetDir, err)
		return
	}

	fetchCmd := exec.Command(gemaidPath, "fetch", url, targetDir)
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		fmt.Printf("ERROR: FETCH FAILED: %v\n%s\n", err, string(output))
		return
	}

	if spec.Purify {
		exec.Command(gemaidPath, "purify", "-root", targetDir).Run()
	}

	if spec.PostInstall != nil {
		if err := spec.PostInstall(toolchainDir); err != nil {
			fmt.Printf("ERROR: POST-INSTALL FAILED: %v\n", err)
			return
		}
	}

	fmt.Println("SUCCESS")
}

func setupPodmanMachine() {
	podmanExe := resolveSovereignTool("podman")
	fmt.Printf("🐳 Initializing fresh Podman Machine (Capped at 50GB)... ")
	cmd := exec.Command(podmanExe, "machine", "init", "--cpus", "4", "--memory", "4096", "--disk-size", "50")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("FAILED or ALREADY EXISTS: %v\n", err)
	} else {
		fmt.Println("SUCCESS")
		fmt.Println("▶️  Starting machine...")
		exec.Command(podmanExe, "machine", "start").Run()
	}
}

func runBinaryNormalizationReport() {
	fmt.Println("\n--- Sovereign Scorch Report ---")
	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && (d.Name() == "81000-Toolchain-External" || d.Name() == "82000-Toolchain-Fleet" || d.Name() == "node_modules" || d.Name() == ".git") {
			return filepath.SkipDir
		}
		ext := filepath.Ext(path)
		if ext == ".exe" || ext == ".cmd" || ext == ".bat" {
			fmt.Printf("⚠️  ORPHAN: %s\n", path)
		}
		return nil
	})
}
