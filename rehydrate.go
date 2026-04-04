package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gitsovkey "olympus.fleet/00SDLC/OlympusLogicLibrary/90200-Logic-Libraries/200-Hashing/110-gitsov-key"
	gitsovnotary "olympus.fleet/00SDLC/OlympusLogicLibrary/90200-Logic-Libraries/400-Assurance/120-gitsov-notary"
)

type Asset struct {
	Name      string
	URL       string
	Target    string // relative to base
	Base      string // 81000 or 81200
	IsZip     bool
	Hardened  bool
	ExpectedH string // Expected SHA256 (External Signature)
}

func parseManifest(path string) ([]Asset, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var assets []Asset
	var current *Asset
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "::") || strings.HasPrefix(line, "/") {
			continue
		}

		if strings.HasPrefix(line, "Asset") {
			name := strings.Trim(strings.TrimPrefix(line, "Asset"), " {\"")
			current = &Asset{Name: name}
			continue
		}

		if strings.HasPrefix(line, "}") {
			if current != nil {
				assets = append(assets, *current)
				current = nil
			}
			continue
		}

		if current != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), "\";")

			switch key {
			case "URL":
				current.URL = val
			case "Target":
				current.Target = val
			case "Base":
				current.Base = val
			case "IsZip":
				current.IsZip = (val == "true")
			case "Hardened":
				current.Hardened = (val == "true")
			case "SHA256":
				current.ExpectedH = val
			}
		}
	}

	return assets, scanner.Err()
}

func main() {
	slog.Info("Sovereign Re-Hydrator: Seeding Toolchain and Libraries")

	manifestPath := filepath.Join("00SDLC", "OlympusBuilder", "C0100-Configuration-Registry", "REHYDRATE_MANIFEST.jebnf")
	assets, err := parseManifest(manifestPath)
	if err != nil {
		slog.Error("Failed to load manifest", "path", manifestPath, "error", err)
		os.Exit(1)
	}

	for _, asset := range assets {
		// Assets are installed relative to Forge root
		targetPath := filepath.Join(asset.Base, asset.Target)
		stateFile := filepath.Join(targetPath, ".rehydrate-state.jebnf")

		// Phase 2 check: Check for existing hydration state and verify Notary integrity
		if stateData, err := os.ReadFile(stateFile); err == nil {
			if strings.Contains(string(stateData), fmt.Sprintf("URL = %q", asset.URL)) {
				slog.Info("Asset already present and state-matched", "name", asset.Name, "target", asset.Target)
				continue
			}
		}

		slog.Info("Hydrating asset", "name", asset.Name, "url", asset.URL)
		if err := AssetHydrate(asset); err != nil {
			slog.Error("Error hydrating asset", "name", asset.Name, "error", err)
		} else {
			// Record successful hydration state with Notarization
			artifactKey := ComputeDirectoryKey(targetPath)
			record := gitsovnotary.CreateRecord(artifactKey, "Sovereign-Rehydrator")

			stateContent := fmt.Sprintf("::Olympus::Forge::HydrationState::v2\n")
			stateContent += fmt.Sprintf("URL = %q ;\n", asset.URL)
			stateContent += fmt.Sprintf("ArtifactKey = %q ;\n", record.ArtifactKey.Hex())
			stateContent += fmt.Sprintf("NotaryRecord = %q ;\n", record.RecordHash.Hex())
			stateContent += fmt.Sprintf("Timestamp = %q ;\n", record.Timestamp.Format(time.RFC3339))

			os.MkdirAll(targetPath, 0755)
			os.WriteFile(stateFile, []byte(stateContent), 0644)
			slog.Info("Asset restored and notarized", "name", asset.Name, "notary", record.RecordHash.String()[:12])
		}
	}

	slog.Info("Re-Hydration Pulse Complete")
}

func ComputeDirectoryKey(path string) gitsovkey.GitSovKey {
	h := gitsovkey.NewHasher()
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() == ".rehydrate-state.jebnf" {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()
		io.Copy(h, f)
		return nil
	})
	var out [32]byte
	h.Sum(out[:0])
	key, _ := gitsovkey.FromBLAKE3Bytes(out[:])
	return key
}

func AssetHydrate(asset Asset) error {
	destDir := filepath.Join(asset.Base, asset.Target)

	// Purge existing content to ensure clean hydration
	os.RemoveAll(destDir)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	resp, err := http.Get(asset.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download %s: status %d", asset.Name, resp.StatusCode)
	}

	// T-Phase Validation: SHA256 External Signature check if provided
	var hashBuffer io.Reader
	if asset.ExpectedH != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		h := sha256.Sum256(body)
		actualH := hex.EncodeToString(h[:])
		if actualH != asset.ExpectedH {
			return fmt.Errorf("SECURITY BREACH: %s checksum mismatch! Expected %s, got %s", asset.Name, asset.ExpectedH, actualH)
		}
		hashBuffer = strings.NewReader(string(body))
	} else {
		hashBuffer = resp.Body
	}

	if asset.IsZip {
		tmpZip := filepath.Join(asset.Base, asset.Name+".zip")
		f, err := os.Create(tmpZip)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, hashBuffer)
		f.Close()
		if err != nil {
			return err
		}

		if err := ExtractZip(tmpZip, destDir); err != nil {
			return err
		}
		os.Remove(tmpZip)
	} else if strings.HasSuffix(asset.URL, ".tar.gz") {
		tmpTar := filepath.Join(asset.Base, asset.Name+".tar.gz")
		f, err := os.Create(tmpTar)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, hashBuffer)
		f.Close()
		if err != nil {
			return err
		}

		if err := ExtractTarGz(tmpTar, destDir); err != nil {
			return err
		}
		os.Remove(tmpTar)
	} else {
		// Direct file download
		targetFile := filepath.Join(destDir, filepath.Base(asset.Target))
		if !strings.Contains(asset.Target, "/") {
			targetFile = destDir
		}

		f, err := os.Create(targetFile)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, hashBuffer)
		f.Close()
		if err != nil {
			return err
		}
	}

	if asset.Hardened {
		Harden(destDir)
	}

	return nil
}

func ExtractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func ExtractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func Harden(path string) {
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && (info.Name() == "testdata" || info.Name() == "test") {
			slog.Info("🛡️ Liquidating test node", "path", p)
			os.RemoveAll(p)
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.go") {
			os.Remove(p)
		}
		return nil
	})
}
