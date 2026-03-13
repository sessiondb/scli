package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/AlecAivazis/survey/v2"
)

const (
	scliReleasesRepo = "sessiondb/scli"
	releasesAPIURL  = "https://api.github.com/repos/" + scliReleasesRepo + "/releases?per_page=5"
)

// ghRelease represents a minimal GitHub release for listing.
type ghRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []ghReleaseAsset `json:"assets"`
}

// ghReleaseAsset represents a release asset.
type ghReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// runUpdate fetches the latest 5 scli releases, prompts the user to select a version,
// then downloads the binary for the current platform and replaces the running executable.
func runUpdate() error {
	releases, err := fetchReleases(5)
	if err != nil {
		return err
	}
	if len(releases) == 0 {
		return fmt.Errorf("no releases found for %s", scliReleasesRepo)
	}

	options := make([]string, len(releases))
	for i, r := range releases {
		options[i] = r.TagName
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select scli version to install:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}

	var chosen *ghRelease
	for i := range releases {
		if releases[i].TagName == selected {
			chosen = &releases[i]
			break
		}
	}
	if chosen == nil {
		return fmt.Errorf("selected version not found: %s", selected)
	}

	assetName := scliAssetName(selected)
	var downloadURL string
	for _, a := range chosen.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no asset %s for release %s (available: %v)", assetName, selected, assetNames(chosen.Assets))
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path: %w", err)
	}

	fmt.Printf("Downloading %s ...\n", assetName)
	if err := downloadAndReplace(downloadURL, exePath); err != nil {
		return err
	}
	fmt.Printf("Updated scli to %s at %s\n", selected, exePath)
	return nil
}

// fetchReleases returns up to limit releases from the GitHub API.
func fetchReleases(limit int) ([]ghRelease, error) {
	req, err := http.NewRequest(http.MethodGet, releasesAPIURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("releases API: %s", resp.Status)
	}

	var releases []ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	if len(releases) > limit {
		releases = releases[:limit]
	}
	return releases, nil
}

// scliAssetName returns the expected asset name for the current GOOS/GOARCH and tag.
// Matches .github/workflows/release.yml: scli-${VERSION}-linux-amd64, ... -windows-amd64.exe
func scliAssetName(tag string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	suffix := ""
	if goos == "windows" {
		suffix = ".exe"
	}
	return fmt.Sprintf("scli-%s-%s-%s%s", tag, goos, goarch, suffix)
}

func assetNames(assets []ghReleaseAsset) []string {
	names := make([]string, len(assets))
	for i := range assets {
		names[i] = assets[i].Name
	}
	return names
}

// downloadAndReplace downloads the URL to a temp file then replaces the executable at exePath.
func downloadAndReplace(downloadURL, exePath string) error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: %s", resp.Status)
	}

	dir := filepath.Dir(exePath)
	f, err := os.CreateTemp(dir, ".scli-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := f.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		return fmt.Errorf("write: %w", err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	// Replace current binary: remove destination then rename (Unix). On Windows, rename over existing may fail.
	if runtime.GOOS == "windows" {
		// Windows: cannot replace running exe. Write to .exe.new and instruct user.
		newPath := exePath + ".new"
		if err := os.Rename(tmpPath, newPath); err != nil {
			return fmt.Errorf("rename to .new: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Downloaded to %s. Replace %s with it and restart (e.g. copy over after closing scli).\n", newPath, exePath)
		return nil
	}

	if err := os.Remove(exePath); err != nil {
		return fmt.Errorf("remove current binary: %w", err)
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	_ = written // optional: could log bytes written
	return nil
}
