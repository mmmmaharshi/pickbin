package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: pickbin <github-release-url>\n")
		os.Exit(1)
	}

	releaseURL := os.Args[1]
	apiURL, tag := parseGitHubURL(releaseURL)
	if apiURL == "" {
		fmt.Fprintf(os.Stderr, "Error: not a valid GitHub release URL\n")
		os.Exit(1)
	}

	assets, err := fetchRelease(apiURL, tag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching release: %v\n", err)
		os.Exit(1)
	}

	best := findBestMatch(assets)
	if best == nil {
		fmt.Println("No matching binary found. Available assets:")
		for _, a := range assets {
			fmt.Printf("  %s (%s)\n", a.Name, formatSize(a.Size))
		}
		os.Exit(1)
	}

	fmt.Printf("✓ Detected: %s / %s\n", runtime.GOOS, normalizeArch(runtime.GOARCH))
	fmt.Printf("\nRecommended:\n  %s\n\n", getDownloadURL(best))
	fmt.Printf("Asset:    %s\n", best.Name)
	fmt.Printf("Size:     %s\n", formatSize(best.Size))
	fmt.Printf("Downloads: %d\n", best.DownloadCount)
	if best.ContentType != "" {
		fmt.Printf("Type:     %s\n", best.ContentType)
	}
}

func parseGitHubURL(raw string) (apiURL string, tag string) {
	// Accept both https://github.com/user/repo/releases/tag/v1.0 and
	// https://api.github.com/repos/user/repo/releases/tags/v1.0
	const apiPrefix = "https://api.github.com/repos/"

	if strings.HasPrefix(raw, apiPrefix) {
		parts := strings.SplitN(raw[len(apiPrefix):], "/", 3)
		if len(parts) == 3 && parts[0] != "" && parts[1] != "" {
			return raw, parts[2]
		}
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", ""
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[len(pathParts)-2] != "tag" {
		return "", ""
	}

	user := pathParts[0]
	repo := pathParts[1]
	tag = pathParts[len(pathParts)-1]

	return fmt.Sprintf("%s%s/%s/releases/tags/%s", apiPrefix, user, repo, tag), tag
}

type Asset struct {
	Name          string `json:"name"`
	URL           string `json:"browser_download_url"`
	Size          int64  `json:"size"`
	ContentType   string `json:"content_type"`
	DownloadCount int    `json:"download_count"`
}

type Release struct {
	Assets []Asset `json:"assets"`
}

func fetchRelease(apiURL, tag string) ([]Asset, error) {
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "pickbin")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return release.Assets, nil
}

func getDownloadURL(a *Asset) string {
	return a.URL
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func hasKeyword(name, keyword string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, strings.ToLower(keyword))
}

func archMatches(assetArch, goArch string) bool {
	goLower := strings.ToLower(goArch)
	aLower := strings.ToLower(assetArch)

	// Direct match
	if goLower == aLower { return true }

	// amd64 synonyms
	if goLower == "amd64" {
		return aLower == "x86_64" || aLower == "x64" || aLower == "amd64"
	}

	// arm64 synonyms
	if goLower == "arm64" {
		return aLower == "aarch64" || aLower == "arm64"
	}

	return false
}

func normalizeArch(a string) string {
	if len(a) < 2 { return strings.ToUpper(a) }
	return strings.ToUpper(a[:1]) + a[1:]
}

func findBestMatch(assets []Asset) *Asset {
	if len(assets) == 0 { return nil }

	goOS := runtime.GOOS
	goArch := runtime.GOARCH

	type candidate struct {
		asset *Asset
		score int
	}

	var candidates []candidate

	for i := range assets {
		a := &assets[i]

		// Strip extension for keyword search
		nameOnly := stripExtension(a.Name)
		nameLower := strings.ToLower(nameOnly)

		// Step 1: Skip obvious non-binaries
		lowerName := strings.ToLower(a.Name)
		if isNonBinary(lowerName) { continue }

		// Step 2: Check for embedded OS/arch in name (e.g., linux-amd64)
		osMatch := hasOS(nameLower, goOS)
		archMatch := hasArch(nameLower, goArch)

		if !osMatch && !archMatch { continue } // skip if neither matches

		score := 0
		if osMatch { score += 10 }
		if archMatch { score += 10 }

		candidates = append(candidates, candidate{asset: a, score: score})
	}

	if len(candidates) == 0 { return nil }

	// Sort by score desc, then download_count desc
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.score > best.score {
			best = c
		} else if c.score == best.score && c.asset.DownloadCount > best.asset.DownloadCount {
			best = c
		}
	}

	return best.asset
}

func isNonBinary(name string) bool {
	lower := strings.ToLower(name)
	// Rejected: checksums, signatures, source distributions, package formats
	rejections := []string{
		".sha", ".md5", ".sha256", ".sha512", // checksums
		".asc", ".sig", ".gpg",              // signatures
		".source", ".src",                    // source tarballs
		"_source",                             // common source suffix
		".deb", ".rpm", ".apk", ".msi",      // package managers, not standalone
	}
	for _, r := range rejections {
		if strings.Contains(lower, r) { return true }
	}
	return false
}

func isArchiveExtension(name string) bool {
	excs := []string{
		".zip", ".tar.gz", ".tgz", ".tar.xz", ".txz",
		".tar.bz2", ".tbz2", ".tar.zst", ".tzst",
		".7z", ".rar", ".lz", ".lrz", ".Z",
	}
	for _, ext := range excs {
		if strings.HasSuffix(name, ext) { return true }
	}
	// Single-char tar extension too
	if strings.HasSuffix(name, ".tar") { return true }
	return false
}

func hasOS(name, goOS string) bool {
	// Common OS names used in releases
	osAliases := map[string][]string{
		"linux":   {"linux"},
		"windows": {"win", "windows", "win32", "win64", "nt"},
		"darwin":  {"macos", "darwin", "osx", "apple-darwin", "darwin"},
	}

	targets, ok := osAliases[goOS]
	if !ok { return false }

	for _, t := range targets {
		if strings.Contains(name, t) { return true }
	}
	return false
}

func hasArch(name, goArch string) bool {
	// Split on common delimiters to get individual words
	words := tokenize(name)

	// Find which word matches this architecture
	for _, w := range words {
		if archMatches(w, goArch) { return true }
	}

	// Also check as substrings (handles x86_64, arm64, etc.)
	substrings := map[string]string{
		"amd64": "amd64",
		"x86_64":"amd64",
		"x64":   "amd64",
		"aarch64":"arm64",
		"arm64": "arm64",
	}

	for assetSubstr, goTarget := range substrings {
		if goArch == goTarget && strings.Contains(name, assetSubstr) {
			return true
		}
	}

	return false
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	// Replace common separators with space
	for _, sep := range []string{"-", "_", "."} {
		s = strings.ReplaceAll(s, sep, " ")
	}
	return strings.Fields(s)
}

func stripExtension(name string) string {
	// Remove known archive extensions first (longer ones)
	excs := []string{
		".tar.gz", ".tar.xz", ".tar.bz2", ".tar.zst",
		".tar.zst", ".tar.zstd",
		".deb", ".rpm", ".apk",
		".msi", ".exe", ".bin",
		".tar", ".zip", ".7z", ".rar",
	}
	for _, ext := range excs {
		if strings.HasSuffix(name, ext) {
			return name[:len(name)-len(ext)]
		}
	}
	// Single extension
	if strings.Contains(name, ".") {
		lastDot := strings.LastIndex(name, ".")
		return name[:lastDot]
	}
	return name
}
