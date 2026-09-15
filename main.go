package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: pickbin <github-repo-or-release-url>\n")
		os.Exit(1)
	}
	apiURL := parseInput(os.Args[1])
	if apiURL == "" {
		fmt.Fprintf(os.Stderr, "Error: not a valid GitHub URL\n")
		os.Exit(1)
	}
	assets, err := fetchRelease(apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching release: %v\n", err)
		os.Exit(1)
	}
	releasePageURL := buildReleasePageURL(apiURL)
	best := findBestMatch(assets, runtime.GOOS, runtime.GOARCH)
	if best == nil {
		fmt.Println("No matching binary found. Available assets:")
		for _, a := range assets {
			fmt.Printf("  %s (%s)\n", a.Name, formatSize(a.Size))
		}
		os.Exit(1)
	}
	fmt.Printf("✓ Detected: %s / %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("\nRecommended:\n  %s\n\n", best.URL)
	fmt.Printf("Asset:    %s\n", best.Name)
	fmt.Printf("Size:     %s\n", formatSize(best.Size))
	fmt.Printf("Downloads: %d\n", best.DownloadCount)
	if best.ContentType != "" {
		fmt.Printf("Type:     %s\n", best.ContentType)
	}
	if err := openBrowser(releasePageURL); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser: %v\n", err)
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
