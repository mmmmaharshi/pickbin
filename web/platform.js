// canonical alias tables — keep in sync with platform.go
// generated from platform.go archMatches/hasOS/isNonBinary
export const archSynonyms = {
  amd64: ["x86_64", "x64"],
  arm64: ["aarch64"],
  "386": ["i386", "i486", "i586", "i686", "x86", "ia32"],
  arm: ["armv5", "armv6", "armv6l", "armv7", "armv7l", "armhf", "armel"],
};

export const osAliases = {
  linux: ["linux"],
  windows: ["win", "windows", "win32", "win64"],
  darwin: ["macos", "darwin", "osx", "apple-darwin"],
};

export const rejections = [
  ".sha", ".md5", ".sha256", ".sha512",
  ".asc", ".sig", ".gpg",
  ".source", ".src", "_source",
  ".deb", ".rpm", ".apk", ".msi",
];

export function archMatches(assetArch, goArch) {
  const a = assetArch.toLowerCase();
  const g = goArch.toLowerCase();
  if (a === g) return true;
  for (const syn of archSynonyms[g] || []) {
    if (a === syn) return true;
  }
  return false;
}

export function tokenize(s) {
  return s.toLowerCase().split(/[-_.]/).filter(Boolean);
}

export function hasOS(name, goOS) {
  const lower = name.toLowerCase();
  for (const t of osAliases[goOS] || []) {
    if (lower.includes(t)) return true;
  }
  return false;
}

export function hasArch(name, goArch) {
  for (const w of tokenize(name)) {
    if (archMatches(w, goArch)) return true;
  }
  return false;
}

export function isNonBinary(name) {
  const lower = name.toLowerCase();
  for (const r of rejections) {
    if (lower.includes(r)) return true;
  }
  return false;
}

export function findBestMatch(assets, goOS, goArch) {
  if (!assets || assets.length === 0) return null;
  let best = null;
  let bestScore = -1;
  for (const a of assets) {
    if (isNonBinary(a.name)) continue;
    const osMatch = hasOS(a.name, goOS);
    const archMatch = hasArch(a.name, goArch);
    if (!osMatch && !archMatch) continue;
    let score = 0;
    if (osMatch) score += 10;
    if (archMatch) score += 10;
    score += Math.min(a.download_count || 0, 1000) / 1000;
    if (score > bestScore) {
      bestScore = score;
      best = a;
    }
  }
  return best;
}
