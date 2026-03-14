package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	frontendAssetName     = "sessiondb-frontend-build.tar.gz"
	backendChecksumsName  = "checksums.txt"
	frontendChecksumsName = "checksums-frontend.txt"
)

// frontendAssetCandidates returns supported frontend asset names in priority order.
func frontendAssetCandidates() []string {
	return []string{
		frontendAssetName,
		fmt.Sprintf("sessiondb-ui-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH),
		fmt.Sprintf("sessiondb-ui-%s-%s", runtime.GOOS, runtime.GOARCH),
	}
}

// findFirstAssetURL returns the first matching asset URL and its asset name.
func findFirstAssetURL(release *githubRelease, candidates []string) (string, string) {
	for _, name := range candidates {
		if url := findAssetURL(release, name); url != "" {
			return url, name
		}
	}
	return "", ""
}

// releaseAssetNames returns all asset names from a release.
func releaseAssetNames(release *githubRelease) []string {
	out := make([]string, 0, len(release.Assets))
	for _, a := range release.Assets {
		out = append(out, a.Name)
	}
	return out
}

// ensureUIBinaryForVersion downloads the UI server binary from the client release into versionDir/ui/sessiondb-ui
// if the release contains sessiondb-ui-<os>-<arch>. Returns (true, nil) if the binary was just added, (false, nil) if already present, error if missing in release.
func ensureUIBinaryForVersion(version string, versionDir string) (added bool, err error) {
	uiBinPath := filepath.Join(versionDir, "ui", uiBinaryName)
	if _, err := os.Stat(uiBinPath); err == nil {
		return false, nil // already present
	}
	version = normalizeTag(version)
	frontendRelease, err := fetchReleaseByTagFrom(sessiondbClientRepo, version)
	if err != nil {
		return false, fmt.Errorf("frontend release %s: %w", version, err)
	}
	uiBinAssetName := fmt.Sprintf("sessiondb-ui-%s-%s", runtime.GOOS, runtime.GOARCH)
	uiBinURL := findAssetURL(frontendRelease, uiBinAssetName)
	if uiBinURL == "" {
		return false, fmt.Errorf("release %s does not contain %s; use a release that includes the UI binary or install with --force", version, uiBinAssetName)
	}
	tmpDir, err := os.MkdirTemp("", "scli-ui-")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(tmpDir)
	uiTmp := filepath.Join(tmpDir, uiBinAssetName)
	if _, err := downloadAsset(uiBinURL, uiTmp); err != nil {
		return false, fmt.Errorf("download UI binary: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(versionDir, "ui"), 0755); err != nil {
		return false, err
	}
	data, err := os.ReadFile(uiTmp)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(uiBinPath, data, 0755); err != nil {
		return false, err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] installed UI server binary to %s\n", uiBinPath)
	}
	return true, nil
}

// get installs the given SessionDB version from GitHub Releases into installRoot.
// It downloads backend from sessiondb/service and frontend from sessiondb/client for the same tag,
// verifies their checksums using the respective checksum files, extracts under versions/<tag>/,
// writes setup.sh and sessiondb.yaml, and updates the current symlink.
// If forceReinstall is true and the version dir already exists, it is removed and reinstalled.
// If the version is already installed and the UI binary is missing, it is downloaded and added.
func get(version string, destDir string, forceReinstall bool) error {
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
		if forceReinstall {
			if err := os.RemoveAll(versionDir); err != nil {
				return fmt.Errorf("remove existing version dir: %w", err)
			}
			if verbose {
				fmt.Fprintf(os.Stderr, "[verbose] removed existing %s for reinstall\n", versionDir)
			}
		} else {
			// Already installed: ensure UI binary if missing
			added, err := ensureUIBinaryForVersion(version, versionDir)
			if err != nil {
				return fmt.Errorf("version %s already installed at %s; %w (use --force to reinstall)", version, versionDir, err)
			}
			if added {
				fmt.Printf("Added UI binary to existing install at %s/ui/%s\n", versionDir, uiBinaryName)
			} else {
				fmt.Printf("Version %s already installed at %s (UI binary present)\n", version, versionDir)
			}
			return nil
		}
	}

	backendName := backendAssetName()
	backendURL := findAssetURL(backendRelease, backendName)
	backendChecksumsURL := findAssetURL(backendRelease, backendChecksumsName)

	frontendURL, frontendAssetSelected := findFirstAssetURL(frontendRelease, frontendAssetCandidates())
	frontendChecksumsURL := findAssetURL(frontendRelease, frontendChecksumsName)

	if backendURL == "" {
		return fmt.Errorf("asset %s not found in release %s of %s (platform %s/%s)", backendName, version, sessiondbServiceRepo, runtime.GOOS, runtime.GOARCH)
	}
	if frontendURL == "" {
		return fmt.Errorf("none of frontend assets %v found in release %s of %s; available: %v", frontendAssetCandidates(), version, sessiondbClientRepo, releaseAssetNames(frontendRelease))
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
	frontendPath := filepath.Join(tmpDir, frontendAssetSelected)
	frontendSHA, err := downloadAsset(frontendURL, frontendPath)
	if err != nil {
		return fmt.Errorf("download frontend: %w", err)
	}
	if frontendChecksumsURL != "" {
		f, _ := os.Open(frontendChecksumsPath)
		checksums, _ := parseChecksums(f)
		f.Close()
		if want, ok := checksums[frontendAssetSelected]; ok && want != "" {
			if err := verifyChecksum(frontendPath, want); err != nil {
				return fmt.Errorf("frontend checksum: %w", err)
			}
		}
	} else if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] no frontend checksums for %s, skipping frontend verification (sha256: %s)\n", frontendAssetSelected, frontendSHA)
	}
	if strings.HasSuffix(frontendAssetSelected, ".tar.gz") {
		if err := extractTarGzip(frontendPath, filepath.Join(versionDir, "ui")); err != nil {
			return fmt.Errorf("extract frontend: %w", err)
		}
		// If release also has UI server binary, download it so "scli run --component ui" and deploy --component ui work.
		uiBinAssetName := fmt.Sprintf("sessiondb-ui-%s-%s", runtime.GOOS, runtime.GOARCH)
		if uiBinURL := findAssetURL(frontendRelease, uiBinAssetName); uiBinURL != "" {
			uiBinPath := filepath.Join(versionDir, "ui", uiBinaryName)
			uiTmp := filepath.Join(tmpDir, uiBinAssetName)
			if _, err := downloadAsset(uiBinURL, uiTmp); err == nil {
				data, _ := os.ReadFile(uiTmp)
				_ = os.WriteFile(uiBinPath, data, 0755)
				if verbose {
					fmt.Fprintf(os.Stderr, "[verbose] installed UI server binary to %s\n", uiBinPath)
				}
			}
		}
	} else {
		// Binary-style frontend artifact (mapped per platform).
		uiBinPath := filepath.Join(versionDir, "ui", uiBinaryName)
		data, err := os.ReadFile(frontendPath)
		if err != nil {
			return fmt.Errorf("read frontend binary: %w", err)
		}
		if err := os.WriteFile(uiBinPath, data, 0755); err != nil {
			return fmt.Errorf("write frontend binary: %w", err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] installed frontend binary to %s\n", uiBinPath)
		}
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
	fmt.Printf("Backend asset:  %s\n", backendName)
	fmt.Printf("Frontend asset: %s\n", frontendAssetSelected)
	return nil
}
