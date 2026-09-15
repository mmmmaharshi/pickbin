package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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

func parseInput(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	const apiPrefix = "https://api.github.com/repos/"
	if strings.HasPrefix(raw, apiPrefix) {
		rest := strings.TrimSuffix(strings.TrimPrefix(raw, apiPrefix), "/")
		parts := strings.Split(rest, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return apiPrefix + rest + "/releases/latest"
		}
		if len(parts) >= 3 && parts[2] == "releases" {
			return apiPrefix + rest
		}
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if u.Host != "github.com" && u.Host != "www.github.com" {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	user, repo := parts[0], parts[1]
	if len(parts) >= 5 && parts[2] == "releases" && parts[3] == "tag" && parts[4] != "" {
		tag := parts[4]
		return fmt.Sprintf("%s%s/%s/releases/tags/%s", apiPrefix, user, repo, tag)
	}
	return fmt.Sprintf("%s%s/%s/releases/latest", apiPrefix, user, repo)
}

func buildReleasePageURL(apiURL string) string {
	const apiPrefix = "https://api.github.com/repos/"
	if !strings.HasPrefix(apiURL, apiPrefix) {
		return apiURL
	}
	rest := strings.TrimPrefix(apiURL, apiPrefix)
	rest = strings.Replace(rest, "/releases/tags/", "/releases/tag/", 1)
	return "https://github.com/" + rest
}

func fetchRelease(apiURL string) ([]Asset, error) {
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
