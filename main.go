package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: pickbin <github-repo-or-release-url>\n")
		os.Exit(1)
	}

	releaseInput := os.Args[1]

	// Try specific release URL first, fall back to plain repo
	apiURL, tag := parseGitHubURL(releaseInput)
	if apiURL == "" {
		apiURL = parseRepoURL(releaseInput)
		tag = "latest"
	}
	if apiURL == "" {
		fmt.Fprintf(os.Stderr, "Error: not a valid GitHub URL\n")
		os.Exit(1)
	}

	assets, err := fetchRelease(apiURL, tag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching release: %v\n", err)
		os.Exit(1)
	}

	// Derive the release page URL for browser opening
	releasePageURL := buildReleasePageURL(apiURL, tag)
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

	// Open release page in browser
	if err := openBrowser(releasePageURL); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser: %v\n", err)
	}
}

func parseGitHubURL(raw string) (apiURL string, tag string) {
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

func parseRepoURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	// Accept: github.com/user/repo OR api.github.com/repos/user/repo
	const apiPrefix = "https://api.github.com/repos/"
	if strings.HasPrefix(raw, apiPrefix) {
		rest := strings.TrimSuffix(raw[len(apiPrefix):], "/")
		parts := strings.Split(rest, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return fmt.Sprintf("%s%s/releases/latest", apiPrefix, rest)
		}
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) < 2 || pathParts[0] == "" || pathParts[1] == "" {
		return ""
	}

	return fmt.Sprintf("%s%s/%s/releases/latest", apiPrefix, pathParts[0], pathParts[1])
}

func buildReleasePageURL(apiURL, tag string) string {
	// Convert API URL to GitHub HTML release page
	// api.github.com/repos/{owner}/{repo}/releases/tags/{tag}
	// → github.com/{owner}/{repo}/releases/tag/{tag}

	const apiPrefix = "https://api.github.com/repos/"
	if !strings.HasPrefix(apiURL, apiPrefix) {
		return apiURL
	}

	rest := strings.TrimPrefix(apiURL, apiPrefix)
	// API uses /tags/, HTML uses /tag/
	rest = strings.Replace(rest, "/releases/tags/", "/releases/tag/", 1)
	return "https://github.com/" + rest
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

	if goLower == aLower { return true }

	synonyms := map[string][]string{
		"amd64": {"x86_64", "x64"},
		"arm64": {"aarch64"},
		"386":   {"i386", "i486", "i586", "i686", "x86", "ia32"},
		"arm":   {"armv5", "armv6", "armv6l", "armv7", "armv7l", "armhf", "armel"},
	}

	for _, syn := range synonyms[goLower] {
		if aLower == syn { return true }
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
	for _, w := range tokenize(name) {
		if archMatches(w, goArch) { return true }
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

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux and others
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
