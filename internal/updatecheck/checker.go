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
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-version"
)

const (
	githubAPIURL       = "https://api.github.com/repos/xinnjie/onekeymap-cli/releases/latest"
	checkInterval      = 48 * time.Hour
	cacheFileName      = "last_update_check"
	requestTimeout     = 5 * time.Second
	updateCheckTimeout = 3 * time.Second

	// UI formatting constants
	notificationBoxWidth   = 80
	notificationBoxPadding = 2

	cyan  = lipgloss.Color("#00D7FF")
	green = lipgloss.Color("#00FF87")
	gold  = lipgloss.Color("#FFD700")
	blue  = lipgloss.Color("#87CEEB")
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

// CheckForUpdateMessage checks if a new version is available and returns a formatted message.
// Returns empty string if no update is available or if check should be skipped.
func (c *Checker) CheckForUpdateMessage(ctx context.Context) string {
	if c.currentVersion == "dev" {
		c.logger.DebugContext(ctx, "Skipping update check for dev version")
		return ""
	}

	if !c.shouldCheck() {
		c.logger.DebugContext(ctx, "Skipping update check, checked recently")
		return ""
	}

	checkCtx, cancel := context.WithTimeout(ctx, updateCheckTimeout)
	defer cancel()

	msg, err := c.checkAndGetMessage(checkCtx)
	if err != nil {
		c.logger.DebugContext(checkCtx, "Update check failed", "error", err)
		return ""
	}

	return msg
}

// CheckForUpdateAsync starts an asynchronous update check and returns a channel
// that will receive the update message. The channel is buffered and will be closed
// after writing the result (or empty string if no update).
func (c *Checker) CheckForUpdateAsync(ctx context.Context) <-chan string {
	resultChan := make(chan string, 1)

	go func() {
		defer close(resultChan)
		msg := c.CheckForUpdateMessage(ctx)
		resultChan <- msg
	}()

	return resultChan
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

func (c *Checker) checkAndGetMessage(ctx context.Context) (string, error) {
	release, err := c.fetchLatestVersion(ctx)
	if err != nil {
		return "", err
	}

	// Update the check time regardless of whether there's a new version
	if err := c.updateCheckTime(); err != nil {
		c.logger.DebugContext(ctx, "failed to update check time", "error", err)
	}

	current, err := version.NewVersion(c.currentVersion)
	if err != nil {
		return "", fmt.Errorf("failed to parse current version: %w", err)
	}

	latest, err := version.NewVersion(release.Version)
	if err != nil {
		return "", fmt.Errorf("failed to parse latest version: %w", err)
	}

	if latest.GreaterThan(current) {
		return c.formatUpdateNotification(c.currentVersion, release.Version, release.HTMLURL), nil
	}

	return "", nil
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

func (c *Checker) formatUpdateNotification(currentVersion, newVersion, releaseURL string) string {
	// Define styles
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cyan).
		Padding(0, notificationBoxPadding).
		Width(notificationBoxWidth)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(green)

	versionStyle := lipgloss.NewStyle().
		Foreground(gold)

	linkStyle := lipgloss.NewStyle().
		Foreground(blue).
		Underline(true).
		Inline(true) // Prevent wrapping for link

	// Build content
	var content strings.Builder

	// Title with emoji
	content.WriteString(titleStyle.Render("🎉 A new version of onekeymap-cli is available!"))
	content.WriteString("\n\n")

	// Version info
	content.WriteString(fmt.Sprintf("📦 Current version: %s\n", versionStyle.Render(currentVersion)))
	content.WriteString(fmt.Sprintf("✨ Latest version:  %s\n", versionStyle.Render(newVersion)))
	content.WriteString("\n")

	// Release notes link (ensure it stays on one line)
	content.WriteString("📝 Release notes:\n")
	content.WriteString("   " + linkStyle.Render(releaseURL) + "\n")
	content.WriteString("\n")

	// Update instructions based on OS
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("🚀 Update instructions:"))
	content.WriteString("\n")
	content.WriteString(getUpdateInstructions(runtime.GOOS, newVersion))

	return boxStyle.Render(content.String())
}

// getUpdateInstructions returns OS-specific update instructions
func getUpdateInstructions(goos, newVersion string) string {
	switch goos {
	case "darwin":
		return "  🍎 macOS:\n" +
			"     brew update && brew upgrade onekeymap-cli"
	case "windows":
		return "  ❏ Windows:\n" +
			"     winget upgrade xinnjie.onekeymap-cli\n" +
			"     # or: scoop update onekeymap-cli"
	case "linux":
		return fmt.Sprintf(`  🐧 Linux:
     # Debian/Ubuntu:
     wget https://github.com/xinnjie/onekeymap-cli/releases/download/v%s/onekeymap-cli_%s_x86_64.deb
     sudo dpkg -i onekeymap-cli_%s_x86_64.deb

     # Fedora/RHEL/CentOS:
     wget https://github.com/xinnjie/onekeymap-cli/releases/download/v%s/onekeymap-cli_%s_x86_64.rpm
     sudo rpm -i onekeymap-cli_%s_x86_64.rpm

     # Alpine:
     wget https://github.com/xinnjie/onekeymap-cli/releases/download/v%s/onekeymap-cli_%s_x86_64.apk
     sudo apk add --allow-untrusted onekeymap-cli_%s_x86_64.apk`,
			newVersion, newVersion, newVersion,
			newVersion, newVersion, newVersion,
			newVersion, newVersion, newVersion)
	default:
		return "  📦 Download from:\n" +
			"     https://github.com/xinnjie/onekeymap-cli/releases/latest"
	}
}
