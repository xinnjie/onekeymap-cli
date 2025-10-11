package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

const (
	githubAPIURL       = "https://api.github.com/repos/xinnjie/onekeymap-cli/releases/latest"
	checkInterval      = 48 * time.Hour
	cacheFileName      = "last_update_check"
	requestTimeout     = 5 * time.Second
	updateCheckTimeout = 3 * time.Second
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Name    string `json:"name"`
	Version string `json:"-"`
}

type Checker struct {
	currentVersion string
	logger         *slog.Logger
	cacheDir       string
	httpClient     *http.Client
}

func New(currentVersion string, logger *slog.Logger) *Checker {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	cacheDir = filepath.Join(cacheDir, "onekeymap-cli")

	return &Checker{
		currentVersion: currentVersion,
		logger:         logger,
		cacheDir:       cacheDir,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// CheckForUpdate checks if a new version is available and prints a message if so.
// This runs asynchronously and will not block the main command execution.
func (c *Checker) CheckForUpdate(ctx context.Context) {
	if c.currentVersion == "dev" {
		c.logger.DebugContext(ctx, "skipping update check for dev version")
		return
	}

	if !c.shouldCheck() {
		c.logger.DebugContext(ctx, "skipping update check, checked recently")
		return
	}

	// Run the check in a goroutine with timeout to avoid blocking
	go func() {
		checkCtx, cancel := context.WithTimeout(ctx, updateCheckTimeout)
		defer cancel()

		if err := c.checkAndNotify(checkCtx); err != nil {
			c.logger.DebugContext(checkCtx, "update check failed", "error", err)
		}
	}()
}

func (c *Checker) shouldCheck() bool {
	cacheFile := filepath.Join(c.cacheDir, cacheFileName)
	info, err := os.Stat(cacheFile)
	if err != nil {
		return true // File doesn't exist, should check
	}

	return time.Since(info.ModTime()) > checkInterval
}

func (c *Checker) updateCheckTime() error {
	if err := os.MkdirAll(c.cacheDir, 0o750); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cacheFile := filepath.Join(c.cacheDir, cacheFileName)
	f, err := os.Create(cacheFile)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer f.Close()

	return nil
}

func (c *Checker) checkAndNotify(ctx context.Context) error {
	release, err := c.fetchLatestVersion(ctx)
	if err != nil {
		return err
	}

	// Update the check time regardless of whether there's a new version
	if err := c.updateCheckTime(); err != nil {
		c.logger.DebugContext(ctx, "failed to update check time", "error", err)
	}

	current, err := version.NewVersion(c.currentVersion)
	if err != nil {
		return fmt.Errorf("failed to parse current version: %w", err)
	}

	latest, err := version.NewVersion(release.Version)
	if err != nil {
		return fmt.Errorf("failed to parse latest version: %w", err)
	}

	if latest.GreaterThan(current) {
		c.printUpdateNotification(release.Version, release.HTMLURL)
	}

	return nil
}

func (c *Checker) fetchLatestVersion(ctx context.Context) (GitHubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIURL, nil)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GitHubRelease{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return GitHubRelease{}, fmt.Errorf("failed to parse response: %w", err)
	}

	release.Version = strings.TrimPrefix(release.TagName, "v")

	return release, nil
}

func (c *Checker) printUpdateNotification(newVersion, releaseURL string) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "╭─────────────────────────────────────────────────────────────╮\n")
	fmt.Fprintf(os.Stderr, "│  A new version of onekeymap-cli is available!              │\n")
	fmt.Fprintf(os.Stderr, "│                                                             │\n")
	fmt.Fprintf(os.Stderr, "│  Current version: %-10s                                │\n", c.currentVersion)
	fmt.Fprintf(os.Stderr, "│  Latest version:  %-10s                                │\n", newVersion)
	fmt.Fprintf(os.Stderr, "│                                                             │\n")
	fmt.Fprintf(os.Stderr, "│  Release notes: %-43s │\n", releaseURL)
	fmt.Fprintf(os.Stderr, "│                                                             │\n")
	fmt.Fprintf(os.Stderr, "│  Update instructions:                                       │\n")
	fmt.Fprintf(os.Stderr, "│    macOS:  brew upgrade onekeymap-cli                       │\n")
	fmt.Fprintf(os.Stderr, "│    Linux:  Download from GitHub releases                    │\n")
	fmt.Fprintf(os.Stderr, "╰─────────────────────────────────────────────────────────────╯\n")
	fmt.Fprintf(os.Stderr, "\n")
}
