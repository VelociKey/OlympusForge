package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	WindowsDownload = "https://ollama.com/download/ollama-windows-amd64.zip"
	DarwinDownload  = "https://ollama.com/download/Ollama-darwin.zip"
	LinuxDownload   = "https://ollama.com/download/ollama-linux-amd64"
	Gemma2Model     = "gemma:2b"
	Gemma3Model     = "gemma3:4b"
)

func main() {
	fmt.Println("🚀 George Sovereign Intelligence Bootstrap")
	fmt.Println("-------------------------------------------")

	// 1. Setup Directories
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	binDir := filepath.Join(baseDir, ".ollama-bin")
	modelsDir := filepath.Join(baseDir, ".ollama-models")

	os.MkdirAll(binDir, 0755)
	os.MkdirAll(modelsDir, 0755)

	// 2. Download Ollama
	binaryPath := filepath.Join(binDir, "ollama")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fmt.Printf("📦 Downloading Ollama for %s...\n", runtime.GOOS)
		if err := downloadOllama(binDir, binaryPath); err != nil {
			log.Fatalf("Download failed: %v", err)
		}
	} else {
		fmt.Println("✅ Ollama binary already exists.")
	}

	// 3. Start Ollama Serve (Background)
	fmt.Println("🧠 Starting Ollama server...")
	cmdServe := exec.Command(binaryPath, "serve")
	cmdServe.Env = append(os.Environ(), "OLLAMA_MODELS="+modelsDir)
	if err := cmdServe.Start(); err != nil {
		log.Fatalf("Failed to start Ollama server: %v", err)
	}
	defer cmdServe.Process.Kill()

	// Wait for server to be ready
	time.Sleep(5 * time.Second)

	// 4. Pull Gemma Models
	models := []string{Gemma2Model, Gemma3Model}
	for _, model := range models {
		fmt.Printf("📥 Pulling model: %s (this may take a few minutes)...\n", model)
		cmdPull := exec.Command(binaryPath, "pull", model)
		cmdPull.Env = append(os.Environ(), "OLLAMA_MODELS="+modelsDir)
		cmdPull.Stdout = os.Stdout
		cmdPull.Stderr = os.Stderr
		if err := cmdPull.Run(); err != nil {
			log.Fatalf("Failed to pull model %s: %v", model, err)
		}
	}

	fmt.Println("\n✨ Bootstrap Complete!")
	fmt.Printf("Engine: %s\n", binaryPath)
	fmt.Printf("Models: %s\n", modelsDir)
	fmt.Println("George is now ready for local intelligence.")
}

func downloadOllama(binDir, targetPath string) error {
	var downloadURL string
	switch runtime.GOOS {
	case "windows":
		downloadURL = WindowsDownload
	case "darwin":
		downloadURL = DarwinDownload
	case "linux":
		downloadURL = LinuxDownload
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if strings.HasSuffix(downloadURL, ".zip") {
		tmpZip := filepath.Join(binDir, "ollama.zip")
		out, err := os.Create(tmpZip)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, resp.Body)
		out.Close()
		if err != nil {
			return err
		}

		if err := unzip(tmpZip, binDir); err != nil {
			return err
		}
		os.Remove(tmpZip)
		return nil
	}

	// Direct binary download (Linux)
	out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Flatten: find the binary and put it in dest
		if strings.HasSuffix(f.Name, "ollama.exe") || strings.HasSuffix(f.Name, "/ollama") || f.Name == "ollama" {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			target := filepath.Join(dest, filepath.Base(f.Name))
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
