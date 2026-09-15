package main

import "strings"

func archMatches(assetArch, goArch string) bool {
	a := strings.ToLower(assetArch)
	g := strings.ToLower(goArch)
	if a == g {
		return true
	}
	synonyms := map[string][]string{
		"amd64": {"x86_64", "x64"},
		"arm64": {"aarch64"},
		"386":   {"i386", "i486", "i586", "i686", "x86", "ia32"},
		"arm":   {"armv5", "armv6", "armv6l", "armv7", "armv7l", "armhf", "armel"},
	}
	for _, syn := range synonyms[g] {
		if a == syn {
			return true
		}
	}
	return false
}

func hasOS(name, goOS string) bool {
	aliases := map[string][]string{
		"linux":   {"linux"},
		"windows": {"win", "windows", "win32", "win64"},
		"darwin":  {"macos", "darwin", "osx", "apple-darwin"},
	}
	lower := strings.ToLower(name)
	for _, t := range aliases[goOS] {
		if strings.Contains(lower, t) {
			return true
		}
	}
	return false
}

func hasArch(name, goArch string) bool {
	for _, w := range tokenize(name) {
		if archMatches(w, goArch) {
			return true
		}
	}
	return false
}

func isNonBinary(name string) bool {
	lower := strings.ToLower(name)
	rejections := []string{
		".sha", ".md5", ".sha256", ".sha512",
		".asc", ".sig", ".gpg",
		".source", ".src", "_source",
		".deb", ".rpm", ".apk", ".msi",
	}
	for _, r := range rejections {
		if strings.Contains(lower, r) {
			return true
		}
	}
	return false
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	for _, sep := range []string{"-", "_", "."} {
		s = strings.ReplaceAll(s, sep, " ")
	}
	return strings.Fields(s)
}
