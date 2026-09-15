package main

import (
	"fmt"
	"strings"
)

func findBestMatch(assets []Asset, goOS, goArch string) *Asset {
	if len(assets) == 0 {
		return nil
	}
	type candidate struct {
		asset *Asset
		score int
	}
	var candidates []candidate
	for i := range assets {
		a := &assets[i]
		lower := strings.ToLower(a.Name)
		if isNonBinary(lower) {
			continue
		}
		osMatch := hasOS(lower, goOS)
		archMatch := hasArch(lower, goArch)
		if !osMatch && !archMatch {
			continue
		}
		score := 0
		if osMatch {
			score += 10
		}
		if archMatch {
			score += 10
		}
		candidates = append(candidates, candidate{asset: a, score: score})
	}
	if len(candidates) == 0 {
		return nil
	}
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
