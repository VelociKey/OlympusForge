package main

import (
	"archive/tar"
	"archive/zip"
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

var Manifest = []Asset{
	{
		Name:     "Podman-Windows",
		URL:      "https://github.com/containers/podman/releases/download/v5.4.0/podman-remote-release-windows_amd64.zip",
		Target:   "podman",
		Base:     "81000-Toolchain-External",
		IsZip:    true,
		Hardened: false,
	},
	{
		Name:     "Trivy-Windows",
		URL:      "https://github.com/aquasecurity/trivy/releases/download/v0.69.5/trivy_0.69.5_windows-64bit.zip",
		Target:   "trivy-windows",
		Base:     "81000-Toolchain-External",
		IsZip:    true,
		Hardened: false,
	},
	{
		Name:     "Trivy-Linux",
		URL:      "https://github.com/aquasecurity/trivy/releases/download/v0.69.5/trivy_0.69.5_linux-64bit.tar.gz",
		Target:   "trivy-linux",
		Base:     "81000-Toolchain-External",
		IsZip:    false,
		Hardened: false,
	},
	{
		Name:     "Dagger-Linux",
		URL:      "https://github.com/dagger/dagger/releases/download/v0.20.3/dagger_v0.20.3_linux_amd64.tar.gz",
		Target:   "dagger-linux",
		Base:     "81000-Toolchain-External",
		IsZip:    false,
		Hardened: false,
	},
	{
		Name:     "Dagger",
		URL:      "https://github.com/dagger/dagger/releases/download/v0.20.3/dagger_v0.20.3_windows_amd64.zip",
		Target:   "dagger",
		Base:     "81000-Toolchain-External",
		IsZip:    true,
		Hardened: false,
	},
	{
		Name:     "Go",
		URL:      "https://go.dev/dl/go1.26.0.windows-amd64.zip",
		Target:   "go",
		Base:     "81000-Toolchain-External",
		IsZip:    true,
		Hardened: false,
	},
	{
		Name:     "Buf",
		URL:      "https://github.com/bufbuild/buf/releases/download/v1.50.0/buf-Linux-x86_64.tar.gz",
		Target:   "buf",
		Base:     "81000-Toolchain-External",
		IsZip:    false,
		Hardened: false,
	},
	{
		Name:     "Gcloud",
		URL:      "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-linux-x86_64.tar.gz",
		Target:   "gcloud",
		Base:     "81000-Toolchain-External",
		IsZip:    false,
		Hardened: false,
	},
	{
		Name:     "Connect-RPC",
		URL:      "https://github.com/connectrpc/connect-go/archive/refs/tags/v1.18.1.zip",
		Target:   "connectrpc/connect-go-1.18.1",
		Base:     "81200-Logic-Libraries",
		IsZip:    true,
		Hardened: true,
	},
	{
		Name:     "Go-QUIC",
		URL:      "https://github.com/quic-go/quic-go/archive/refs/tags/v0.50.0.zip",
		Target:   "quic-go/quic-go-0.50.0",
		Base:     "81200-Logic-Libraries",
		IsZip:    true,
		Hardened: true,
	},
}

func main() {
	fmt.Println("🚀 Sovereign Re-Hydrator: Seeding Toolchain and Libraries")

	for _, asset := range Manifest {
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
	os.MkdirAll(destDir, 0755)

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
