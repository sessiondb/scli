package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// sessiondbServiceRepo is the GitHub repo for SessionDB backend releases.
	sessiondbServiceRepo = "sessiondb/service"
	// sessiondbClientRepo is the GitHub repo for SessionDB UI releases.
	sessiondbClientRepo = "sessiondb/client"

	githubReleasesTagURL    = "https://api.github.com/repos/%s/releases/tags/%s"
	githubReleasesLatestURL = "https://api.github.com/repos/%s/releases/latest"
)

// githubRelease represents a GitHub release for asset listing.
type githubRelease struct {
	TagName string               `json:"tag_name"`
	Assets  []githubReleaseAsset `json:"assets"`
}

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// backendAssetName returns the SessionDB backend tarball name for the current platform.
func backendAssetName() string {
	return fmt.Sprintf("sessiondb-backend-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
}

// fetchReleaseByTagFrom returns the release for the given tag (e.g. v1.0.1 or 1.0.1) from the given repo.
func fetchReleaseByTagFrom(repo, tag string) (*githubRelease, error) {
	tag = normalizeTag(tag)
	url := fmt.Sprintf(githubReleasesTagURL, repo, tag)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch release for %s: %w", repo, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %s not found for %s", tag, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release API for %s: %s", repo, resp.Status)
	}
	var r githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// fetchLatestReleaseFrom returns the latest release (by tag) from the given repo.
func fetchLatestReleaseFrom(repo string) (*githubRelease, error) {
	url := fmt.Sprintf(githubReleasesLatestURL, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release for %s: %w", repo, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("latest release API for %s: %s", repo, resp.Status)
	}
	var r githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// fetchReleaseByTag is kept for callers that only care about the backend (service) repo.
func fetchReleaseByTag(tag string) (*githubRelease, error) {
	return fetchReleaseByTagFrom(sessiondbServiceRepo, tag)
}

// fetchLatestRelease is kept for callers that only care about the backend (service) repo.
func fetchLatestRelease() (*githubRelease, error) {
	return fetchLatestReleaseFrom(sessiondbServiceRepo)
}

func normalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag != "" && !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}
	return tag
}

// findAssetURL returns the download URL for the given asset name.
func findAssetURL(release *githubRelease, name string) string {
	for _, a := range release.Assets {
		if a.Name == name {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

// downloadAsset streams the asset to destPath and returns its SHA256 hex digest.
func downloadAsset(url string, destPath string) (sha256Hex string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download: %s", resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", err
	}
	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	h := sha256.New()
	w := io.MultiWriter(out, h)
	n, err := io.Copy(w, resp.Body)
	if err != nil {
		os.Remove(destPath)
		return "", err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] downloaded %s (%d bytes)\n", filepath.Base(destPath), n)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// parseChecksums reads a sha256sum-style file (hex  filename or hex *filename) and returns a map of filename -> hex.
func parseChecksums(r io.Reader) (map[string]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "hex  filename" or "hex *filename"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		hexStr := fields[0]
		name := fields[len(fields)-1]
		// Strip * if present (binary mode indicator)
		if strings.HasPrefix(name, "*") {
			name = name[1:]
		}
		m[name] = hexStr
	}
	return m, nil
}

// verifyChecksum returns nil if file at path has the expected SHA256 (hex).
func verifyChecksum(path string, expectedHex string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != strings.ToLower(strings.TrimSpace(expectedHex)) {
		return fmt.Errorf("checksum mismatch for %s: got %s, want %s", filepath.Base(path), got, expectedHex)
	}
	return nil
}

// extractTarGzip extracts tar.gz from srcPath into destDir. Only paths under destDir are written; path traversal is rejected.
func extractTarGzip(srcPath, destDir string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := h.Name
		if filepath.IsAbs(name) {
			name = strings.TrimPrefix(filepath.Clean(name), "/")
		}
		name = filepath.Clean(name)
		if name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			continue
		}
		target := filepath.Join(destDir, name)
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(targetAbs, destAbs) {
			continue
		}
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			w, err := os.Create(target)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, tr)
			w.Close()
			if err != nil {
				return err
			}
			if h.FileInfo().Mode()&0111 != 0 {
				_ = os.Chmod(target, 0755)
			}
		}
	}
	return nil
}

// getInstallRoot returns the root directory for versions/ and current symlink.
func getInstallRoot(overrideDir string) string {
	if overrideDir != "" {
		return overrideDir
	}
	if root := os.Getenv("SESSIONDB_INSTALL_ROOT"); root != "" {
		return root
	}
	// Prefer /opt/sessiondb if we can write (e.g. root or existing writable dir).
	if os.Getuid() == 0 {
		return "/opt/sessiondb"
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		return filepath.Join(home, ".local", "share", "sessiondb")
	}
	return filepath.Join(".", "sessiondb-install")
}
