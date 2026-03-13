package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	frontendAssetName       = "sessiondb-frontend-build.tar.gz"
	backendChecksumsName    = "checksums.txt"
	frontendChecksumsName   = "checksums-frontend.txt"
)

// get installs the given SessionDB version from GitHub Releases into installRoot.
// It downloads backend from sessiondb/service and frontend from sessiondb/client for the same tag,
// verifies their checksums using the respective checksum files, extracts under versions/<tag>/,
// writes setup.sh and sessiondb.yaml, and updates the current symlink.
func get(version string, destDir string) error {
	installRoot := getInstallRoot(destDir)
	if installRoot != destDir {
		// destDir was explicitly provided (e.g. get v1.0.1 .) — use it as install root
		installRoot = destDir
	}
	installRoot, _ = filepath.Abs(installRoot)

	var backendRelease *githubRelease
	var frontendRelease *githubRelease
	var err error
	if version == "" || strings.ToLower(version) == "latest" {
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] fetching latest release\n")
		}
		backendRelease, err = fetchLatestReleaseFrom(sessiondbServiceRepo)
		if err != nil {
			return err
		}
		version = backendRelease.TagName
		frontendRelease, err = fetchReleaseByTagFrom(sessiondbClientRepo, version)
		if err != nil {
			return fmt.Errorf("frontend release %s not found in %s: %w", version, sessiondbClientRepo, err)
		}
	} else {
		version = normalizeTag(version)
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] fetching release %s\n", version)
		}
		backendRelease, err = fetchReleaseByTagFrom(sessiondbServiceRepo, version)
		if err != nil {
			return fmt.Errorf("backend release %s not found in %s: %w", version, sessiondbServiceRepo, err)
		}
		frontendRelease, err = fetchReleaseByTagFrom(sessiondbClientRepo, version)
		if err != nil {
			return fmt.Errorf("frontend release %s not found in %s: %w", version, sessiondbClientRepo, err)
		}
	}

	versionDir := filepath.Join(installRoot, "versions", version)
	if _, err := os.Stat(versionDir); err == nil {
		return fmt.Errorf("version %s already installed at %s (remove it first to reinstall)", version, versionDir)
	}

	backendName := backendAssetName()
	backendURL := findAssetURL(backendRelease, backendName)
	backendChecksumsURL := findAssetURL(backendRelease, backendChecksumsName)

	frontendURL := findAssetURL(frontendRelease, frontendAssetName)
	frontendChecksumsURL := findAssetURL(frontendRelease, frontendChecksumsName)

	if backendURL == "" {
		return fmt.Errorf("asset %s not found in release %s of %s (platform %s/%s)", backendName, version, sessiondbServiceRepo, runtime.GOOS, runtime.GOARCH)
	}
	if frontendURL == "" {
		return fmt.Errorf("asset %s not found in release %s of %s", frontendAssetName, version, sessiondbClientRepo)
	}

	// Create version dir and subdirs
	if err := os.MkdirAll(filepath.Join(versionDir, "server"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(versionDir, "ui"), 0755); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "scli-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	backendChecksumsPath := filepath.Join(tmpDir, backendChecksumsName)
	if backendChecksumsURL != "" {
		if _, err := downloadAsset(backendChecksumsURL, backendChecksumsPath); err != nil {
			return fmt.Errorf("download backend checksums: %w", err)
		}
	}
	frontendChecksumsPath := filepath.Join(tmpDir, frontendChecksumsName)
	if frontendChecksumsURL != "" {
		if _, err := downloadAsset(frontendChecksumsURL, frontendChecksumsPath); err != nil {
			return fmt.Errorf("download frontend checksums: %w", err)
		}
	}

	// Download backend
	backendPath := filepath.Join(tmpDir, backendName)
	backendSHA, err := downloadAsset(backendURL, backendPath)
	if err != nil {
		return fmt.Errorf("download backend: %w", err)
	}
	if backendChecksumsURL != "" {
		f, _ := os.Open(backendChecksumsPath)
		checksums, _ := parseChecksums(f)
		f.Close()
		if want, ok := checksums[backendName]; ok && want != "" {
			if err := verifyChecksum(backendPath, want); err != nil {
				return fmt.Errorf("backend checksum: %w", err)
			}
		}
	} else if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] no backend checksums, skipping backend verification (sha256: %s)\n", backendSHA)
	}
	if err := extractTarGzip(backendPath, filepath.Join(versionDir, "server")); err != nil {
		return fmt.Errorf("extract backend: %w", err)
	}

	// Download frontend
	frontendPath := filepath.Join(tmpDir, frontendAssetName)
	frontendSHA, err := downloadAsset(frontendURL, frontendPath)
	if err != nil {
		return fmt.Errorf("download frontend: %w", err)
	}
	if frontendChecksumsURL != "" {
		f, _ := os.Open(frontendChecksumsPath)
		checksums, _ := parseChecksums(f)
		f.Close()
		if want, ok := checksums[frontendAssetName]; ok && want != "" {
			if err := verifyChecksum(frontendPath, want); err != nil {
				return fmt.Errorf("frontend checksum: %w", err)
			}
		}
	} else if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] no frontend checksums, skipping frontend verification (sha256: %s)\n", frontendSHA)
	}
	if err := extractTarGzip(frontendPath, filepath.Join(versionDir, "ui")); err != nil {
		return fmt.Errorf("extract frontend: %w", err)
	}

	// Write setup.sh and sessiondb.yaml
	setupPath := filepath.Join(versionDir, "setup.sh")
	if err := os.WriteFile(setupPath, []byte(setupScriptContent), 0755); err != nil {
		return err
	}
	yamlPath := filepath.Join(versionDir, "sessiondb.yaml")
	if err := os.WriteFile(yamlPath, []byte(sessiondbYAMLContent), 0644); err != nil {
		return err
	}

	// Update current symlink
	currentPath := filepath.Join(installRoot, "current")
	_ = os.Remove(currentPath)
	if err := os.Symlink(filepath.Join("versions", version), currentPath); err != nil {
		return fmt.Errorf("create current symlink: %w", err)
	}

	fmt.Printf("Installed %s to %s\n", version, versionDir)
	fmt.Printf("current -> versions/%s\n", version)
	return nil
}
