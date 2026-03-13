package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// releasesRepo is the GitHub repo that holds releases/<version>/binaries (sessiondb/.github).
const releasesRepo = "sessiondb/.github"

// get downloads the repo archive (main branch), extracts releases/<version>/binaries/ into destDir/sessiondb/.
// Version can be with or without "v" (e.g. v0.0.1 or 0.0.1); both path forms are tried (releases/v0.0.1/binaries/ and releases/0.0.1/binaries/).
func get(version string, destDir string) error {
	versionNorm := strings.TrimPrefix(version, "v")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	url := fmt.Sprintf("https://github.com/%s/archive/refs/heads/main.zip", releasesRepo)
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] fetching %s\n", url)
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] response status: %s\n", resp.Status)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	zipPath := filepath.Join(destDir, "repo.zip")
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] downloaded %d bytes to %s\n", written, zipPath)
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	fileNames := make([]string, 0, len(r.File))
	for _, f := range r.File {
		fileNames = append(fileNames, f.Name)
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] zip contains %d entries\n", len(fileNames))
		// Show paths that look like releases/
		var releasePaths []string
		for _, n := range fileNames {
			if strings.Contains(n, "releases/") {
				releasePaths = append(releasePaths, n)
			}
		}
		if len(releasePaths) > 0 {
			fmt.Fprintf(os.Stderr, "[verbose] paths under releases/: %v\n", releasePaths)
		} else {
			fmt.Fprintf(os.Stderr, "[verbose] no paths containing 'releases/' found; first 5 entries: %v\n", firstN(fileNames, 5))
		}
	}
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}
	sessiondbDir := filepath.Join(destDir, "sessiondb")
	// Try both version forms: repo may use "releases/v0.0.1/binaries/" or "releases/0.0.1/binaries/"
	needles := []string{
		"releases/" + version + "/binaries/",   // user passed v0.0.1 -> releases/v0.0.1/binaries/
		"releases/" + versionNorm + "/binaries/", // releases/0.0.1/binaries/
	}
	if version == versionNorm {
		needles = []string{"releases/" + version + "/binaries/"}
	}
	var extracted int
	var usedNeedle string
	for _, needle := range needles {
		extracted = 0
		usedNeedle = needle
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] trying path prefix: %q\n", needle)
		}
		for _, f := range r.File {
			idx := strings.Index(f.Name, needle)
			if idx < 0 {
				continue
			}
			rel := f.Name[idx+len(needle):]
			if rel == "" {
				continue
			}
			rel = strings.TrimSuffix(rel, "/")
			if rel == "" {
				continue
			}
			dst := filepath.Join(sessiondbDir, rel)
			dstAbs, err := filepath.Abs(dst)
			if err != nil {
				continue
			}
			relToDest, err := filepath.Rel(destAbs, dstAbs)
			if err != nil || strings.HasPrefix(relToDest, "..") {
				continue
			}
			if f.FileInfo().IsDir() {
				_ = os.MkdirAll(dst, 0755)
				extracted++
				if verbose {
					fmt.Fprintf(os.Stderr, "[verbose]   dir: %s\n", rel)
				}
				continue
			}
			_ = os.MkdirAll(filepath.Dir(dst), 0755)
			rc, err := f.Open()
			if err != nil {
				continue
			}
			w, err := os.Create(dst)
			if err != nil {
				rc.Close()
				continue
			}
			_, err = io.Copy(w, rc)
			w.Close()
			rc.Close()
			if err != nil {
				continue
			}
			if f.Mode()&0111 != 0 {
				_ = os.Chmod(dst, 0755)
			}
			extracted++
			if verbose {
				fmt.Fprintf(os.Stderr, "[verbose]   file: %s\n", rel)
			}
		}
		if extracted > 0 {
			break
		}
	}
	_ = os.Remove(zipPath)
	if extracted == 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] no files matched any of %v\n", needles)
		}
		return fmt.Errorf("no binaries found for version %s at %s (run release workflow first)", version, releasesRepo)
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] extracted %d entries using prefix %q\n", extracted, usedNeedle)
	}
	fmt.Printf("Downloaded %s to %s/sessiondb\n", version, destDir)
	return nil
}

func firstN(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
