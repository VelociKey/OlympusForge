package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Asset struct {
	Name     string
	URL      string
	Target   string // relative to base
	Base     string // 81000 or 81200
	IsZip    bool
	Hardened bool
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
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "::") {
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
			}
		}
	}

	return assets, scanner.Err()
}

func main() {
	fmt.Println("🚀 Sovereign Re-Hydrator: Seeding Toolchain and Libraries")

	// Point to the central manifest in the Builder workspace
	manifestPath := filepath.Join("..", "OlympusBuilder", "C0100-Configuration-Registry", "REHYDRATE_MANIFEST.jebnf")
	assets, err := parseManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load manifest at %s: %v\n", manifestPath, err)
		os.Exit(1)
	}

	for _, asset := range assets {
		// Assets are installed relative to Forge root
		targetPath := filepath.Join(asset.Base, asset.Target)

		// If it's a directory, check if it's "real" (contains files)
		if info, err := os.Stat(targetPath); err == nil {
			if info.IsDir() {
				files, _ := os.ReadDir(targetPath)
				if len(files) > 0 {
					fmt.Printf("✔ %s already present. Skipping.\n", asset.Name)
					continue
				}
				os.RemoveAll(targetPath)
			} else {
				fmt.Printf("✔ %s already present. Skipping.\n", asset.Name)
				continue
			}
		}

		fmt.Printf("📦 Hydrating %s...\n", asset.Name)
		if err := AssetHydrate(asset); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error hydrating %s: %v\n", asset.Name, err)
		} else {
			fmt.Printf("✅ %s Restored.\n", asset.Name)
		}
	}

	fmt.Println("---\nRe-Hydration Pulse Complete.")
}

func AssetHydrate(asset Asset) error {
	destDir := filepath.Join(asset.Base, asset.Target)
	
	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
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

	if asset.IsZip {
		tmpZip := filepath.Join(asset.Base, asset.Name+".zip")
		f, err := os.Create(tmpZip)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, resp.Body)
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
		_, err = io.Copy(f, resp.Body)
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
		f, err := os.Create(destDir)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, resp.Body)
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
			fmt.Printf("🛡️ Liquidating test node: %s\n", p)
			os.RemoveAll(p)
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.go") {
			os.Remove(p)
		}
		return nil
	})
}
