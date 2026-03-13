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
func get(version string, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	// GitHub archive of default branch (binaries are pushed to main after tag workflow runs)
	url := fmt.Sprintf("https://github.com/%s/archive/refs/heads/main.zip", releasesRepo)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	zipPath := filepath.Join(destDir, "repo.zip")
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}
	// Zip has one top-level dir like "sessiondb-.github-main/"; we want releases/<version>/binaries/*
	needle := "releases/" + version + "/binaries/"
	sessiondbDir := filepath.Join(destDir, "sessiondb")
	var extracted int
	for _, f := range r.File {
		idx := strings.Index(f.Name, needle)
		if idx < 0 {
			continue
		}
		rel := f.Name[idx+len(needle):]
		if rel == "" {
			continue
		}
		// Trim trailing slash for dirs
		rel = strings.TrimSuffix(rel, "/")
		if rel == "" {
			continue
		}
		dst := filepath.Join(sessiondbDir, rel)
		dstAbs, err := filepath.Abs(dst)
		if err != nil {
			return err
		}
		relToDest, err := filepath.Rel(destAbs, dstAbs)
		if err != nil || strings.HasPrefix(relToDest, "..") {
			continue
		}
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(dst, 0755)
			extracted++
			continue
		}
		_ = os.MkdirAll(filepath.Dir(dst), 0755)
		rc, err := f.Open()
		if err != nil {
			return err
		}
		w, err := os.Create(dst)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(w, rc)
		w.Close()
		rc.Close()
		if err != nil {
			return err
		}
		if f.Mode()&0111 != 0 {
			_ = os.Chmod(dst, 0755)
		}
		extracted++
	}
	_ = os.Remove(zipPath)
	if extracted == 0 {
		return fmt.Errorf("no binaries found for version %s at %s (run release workflow first)", version, releasesRepo)
	}
	fmt.Printf("Downloaded %s to %s/sessiondb\n", version, destDir)
	return nil
}
