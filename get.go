package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	frontendAssetName = "sessiondb-frontend-build.tar.gz"
	checksumsName     = "checksums.txt"
)

// get installs the given SessionDB version from GitHub Releases (sessiondb/service) into installRoot.
// It downloads backend and frontend tarballs, verifies checksums, extracts under versions/<tag>/, writes setup.sh and sessiondb.yaml, and updates the current symlink.
func get(version string, destDir string) error {
	installRoot := getInstallRoot(destDir)
	if installRoot != destDir {
		// destDir was explicitly provided (e.g. get v1.0.1 .) — use it as install root
		installRoot = destDir
	}
	installRoot, _ = filepath.Abs(installRoot)

	var release *githubRelease
	var err error
	if version == "" || strings.ToLower(version) == "latest" {
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] fetching latest release\n")
		}
		release, err = fetchLatestRelease()
		if err != nil {
			return err
		}
		version = release.TagName
	} else {
		version = normalizeTag(version)
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] fetching release %s\n", version)
		}
		release, err = fetchReleaseByTag(version)
		if err != nil {
			return err
		}
	}

	versionDir := filepath.Join(installRoot, "versions", version)
	if _, err := os.Stat(versionDir); err == nil {
		return fmt.Errorf("version %s already installed at %s (remove it first to reinstall)", version, versionDir)
	}

	backendName := backendAssetName()
	backendURL := findAssetURL(release, backendName)
	frontendURL := findAssetURL(release, frontendAssetName)
	checksumsURL := findAssetURL(release, checksumsName)

	if backendURL == "" {
		return fmt.Errorf("asset %s not found in release %s (platform %s/%s)", backendName, version, runtime.GOOS, runtime.GOARCH)
	}
	if frontendURL == "" {
		return fmt.Errorf("asset %s not found in release %s", frontendAssetName, version)
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

	checksumsPath := filepath.Join(tmpDir, checksumsName)
	if checksumsURL != "" {
		if _, err := downloadAsset(checksumsURL, checksumsPath); err != nil {
			return fmt.Errorf("download checksums: %w", err)
		}
	}

	// Download backend
	backendPath := filepath.Join(tmpDir, backendName)
	backendSHA, err := downloadAsset(backendURL, backendPath)
	if err != nil {
		return fmt.Errorf("download backend: %w", err)
	}
	if checksumsURL != "" {
		f, _ := os.Open(checksumsPath)
		checksums, _ := parseChecksums(f)
		f.Close()
		if want, ok := checksums[backendName]; ok && want != "" {
			if err := verifyChecksum(backendPath, want); err != nil {
				return fmt.Errorf("backend checksum: %w", err)
			}
		}
	} else if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] no checksums.txt, skipping backend verification (sha256: %s)\n", backendSHA)
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
	if checksumsURL != "" {
		f, _ := os.Open(checksumsPath)
		checksums, _ := parseChecksums(f)
		f.Close()
		if want, ok := checksums[frontendAssetName]; ok && want != "" {
			if err := verifyChecksum(frontendPath, want); err != nil {
				return fmt.Errorf("frontend checksum: %w", err)
			}
		}
	} else if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] no checksums.txt, skipping frontend verification (sha256: %s)\n", frontendSHA)
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
