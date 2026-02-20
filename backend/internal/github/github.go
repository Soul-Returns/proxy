package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	githubRepo = "Soul-Returns/proxy"
	githubAPI  = "https://api.github.com/repos/" + githubRepo + "/releases"
)

// Release represents a GitHub release
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	CurrentVersion string    `json:"current_version"`
	LatestVersion  string    `json:"latest_version"`
	UpdateChannel  string    `json:"update_channel"`
	Available      bool      `json:"available"`
	Release        *Release  `json:"release,omitempty"`
	CheckedAt      time.Time `json:"checked_at"`
	Error          string    `json:"error,omitempty"`
}

// CheckForUpdates queries GitHub for the latest release based on the update channel
func CheckForUpdates(currentVersion string, channel string) (*UpdateInfo, error) {
	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		UpdateChannel:  channel,
		CheckedAt:      time.Now(),
	}

	// Fetch releases from GitHub
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", githubAPI, nil)
	if err != nil {
		info.Error = err.Error()
		return info, err
	}

	// Add User-Agent header (GitHub API requires it)
	req.Header.Set("User-Agent", "DevProxy-Backend/"+currentVersion)

	resp, err := client.Do(req)
	if err != nil {
		info.Error = fmt.Sprintf("failed to connect to GitHub: %v", err)
		return info, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		info.Error = err.Error()
		return info, err
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		info.Error = fmt.Sprintf("failed to parse GitHub response: %v", err)
		return info, err
	}

	// Find the latest release based on channel preference
	var latestRelease *Release
	for i := range releases {
		r := &releases[i]
		// Skip drafts
		if r.Draft {
			continue
		}

		// For release channel, skip pre-releases
		if channel == "release" && r.Prerelease {
			continue
		}

		// Found the latest applicable release
		latestRelease = r
		break
	}

	if latestRelease == nil {
		info.Error = "no releases found"
		return info, fmt.Errorf("no releases found for channel %s", channel)
	}

	info.LatestVersion = strings.TrimPrefix(latestRelease.TagName, "v")
	info.Release = latestRelease

	// Compare versions
	info.Available = CompareVersions(info.LatestVersion, currentVersion) > 0

	return info, nil
}

// CompareVersions compares two semantic version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func CompareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Simple semantic version comparison
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Pad to same length
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for len(parts1) < maxLen {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < maxLen {
		parts2 = append(parts2, "0")
	}

	// Compare each part
	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		fmt.Sscanf(parts1[i], "%d", &n1)
		fmt.Sscanf(parts2[i], "%d", &n2)

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	return 0
}
