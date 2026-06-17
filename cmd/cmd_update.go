package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/version"
)

const (
	lycheeRepo   = "MD-Mushfiqur123/lychee"
	lycheeGitHub = "github.com/MD-Mushfiqur123/lychee"
)

// GitHubRelease represents the JSON structure from the GitHub releases API.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

func NewUpdateCmd() *cobra.Command {
	var (
		checkOnly    bool
		useBinary    bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Lychee to the latest version",
		Long: `Check for and install the latest version of Lychee.

By default, this command uses 'go install' to update Lychee.
Use --binary to download a pre-built binary instead.
Use --check to only see if an update is available without installing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentVersion := getCurrentVersion()
			latestRelease, err := getLatestRelease()
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			latestVersion := strings.TrimPrefix(latestRelease.TagName, "v")

			if currentVersion == latestVersion {
				fmt.Printf("✅ Lychee is up to date (%s)\n", currentVersion)
				return nil
			}

			fmt.Printf("Current: %s\n", currentVersion)
			fmt.Printf("Latest:  %s\n\n", latestVersion)

			if checkOnly {
				fmt.Printf("An update is available: %s → %s\n", currentVersion, latestVersion)
				fmt.Println("Run 'lychee update' to install the latest version.")
				return nil
			}

			fmt.Println("Updating...")

			if useBinary {
				return downloadBinary(latestRelease)
			}

			return updateViaGoInstall(latestVersion)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Check for updates without installing")
	cmd.Flags().BoolVar(&useBinary, "binary", false, "Download a pre-built binary instead of using go install")

	return cmd
}

// getCurrentVersion returns the current Lychee version.
func getCurrentVersion() string {
	if version.Version == "" || version.Version == "0.0.0" {
		return "dev"
	}
	return version.Version
}

// getLatestRelease fetches the latest release from GitHub.
func getLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", lycheeRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "lychee-cli")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	return &release, nil
}

// updateViaGoInstall uses go install to update Lychee to the latest version.
func updateViaGoInstall(version string) error {
	updateCmd := exec.Command("go", "install", lycheeGitHub+"@latest")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("go install failed: %w\nTry 'lychee update --binary' to download a pre-built binary instead.", err)
	}

	_ = version // version info already shown before install
	fmt.Printf("✅ Updated to %s\n", version)
	return nil
}

// downloadBinary downloads a pre-built binary from GitHub releases.
func downloadBinary(release *GitHubRelease) error {
	assetName := buildAssetName()

	// Find matching asset
	var downloadURL string
	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, assetName) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		// Print available assets for debugging
		fmt.Println("Available assets:")
		for _, asset := range release.Assets {
			fmt.Printf("  - %s\n", asset.Name)
		}
		return fmt.Errorf("no pre-built binary found for %s/%s (looked for: %s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	fmt.Printf("Downloading %s...\n", assetName)

	// Find the current lychee executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	// Download to temp file first
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	tmpPath := exePath + ".new"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write download: %w", err)
	}
	f.Close()

	// On Windows, we need to handle the replacement carefully
	// since the executable is running, we can't directly overwrite it
	if runtime.GOOS == "windows" {
		oldPath := exePath + ".old"
		// Remove existing .old if present
		os.Remove(oldPath)

		// Rename current exe to .old, move new to current
		if err := os.Rename(exePath, oldPath); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to replace executable: %w", err)
		}
		if err := os.Rename(tmpPath, exePath); err != nil {
			// Try to restore the original
			os.Rename(oldPath, exePath)
			os.Remove(tmpPath)
			return fmt.Errorf("failed to replace executable: %w", err)
		}
		// Remove old version (best effort)
		os.Remove(oldPath)
	} else {
		if err := os.Rename(tmpPath, exePath); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to replace executable: %w", err)
		}
		if err := os.Chmod(exePath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	fmt.Printf("✅ Updated to %s\n", strings.TrimPrefix(release.TagName, "v"))
	return nil
}

// buildAssetName constructs the expected asset name for the current platform.
func buildAssetName() string {
	arch := runtime.GOARCH
	osName := runtime.GOOS

	// Normalize architecture names
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "i386"
	}

	ext := ".tar.gz"
	if osName == "windows" {
		ext = ".zip"
	}

	return fmt.Sprintf("lychee-%s-%s%s", osName, arch, ext)
}

// FindLycheeBinary attempts to locate the lychee binary in PATH or common locations.
func FindLycheeBinary() (string, error) {
	// First, check the current executable
	if exe, err := os.Executable(); err == nil {
		if path, err := filepath.EvalSymlinks(exe); err == nil {
			return path, nil
		}
	}

	// Then check PATH
	path, err := exec.LookPath("lychee")
	if err != nil {
		return "", fmt.Errorf("lychee not found in PATH: %w", err)
	}

	return filepath.Abs(path)
}
