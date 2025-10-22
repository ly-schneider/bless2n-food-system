#!/usr/bin/env bash
set -euo pipefail

# Fixed repo as requested
OWNER_REPO="ly-schneider/bless2n-food-system"

# Args:
#   $1 = dest dir (optional, default: ./dist)
#   $2 = tag (optional; if omitted, use latest)
DEST_DIR="${1:-./dist}"
TAG_OPT="${2:-}"   # e.g. v1.2.3; empty = latest
GH_HOST="${GH_HOST:-github.com}"

# Optional strong check: expected signer fingerprint (SHA-256) for the APK
# Set via env: EXPECTED_FINGERPRINT="AA:BB:...:ZZ"  (colons optional)
EXPECTED_FINGERPRINT="${EXPECTED_FINGERPRINT:-}"

mkdir -p "$DEST_DIR"

ensure_gh_auth() {
  if gh auth status -h "$GH_HOST" >/dev/null 2>&1; then
    echo "gh is already authenticated for $GH_HOST"
    return
  fi

  local TOKEN="${GH_TOKEN:-${GITHUB_TOKEN:-}}"
  if [ -z "${TOKEN:-}" ]; then
    echo "ERROR: No GH_TOKEN/GITHUB_TOKEN found. Export a PAT with repo read access." >&2
    echo "       Fine-grained PAT: grant Contents: Read for this repo; Classic PAT: 'repo' scope." >&2
    exit 1
  fi

  echo "Logging in gh with token from env for $GH_HOST…"
  printf '%s' "$TOKEN" | gh auth login --hostname "$GH_HOST" --with-token
}

echo "→ Ensuring GitHub auth…"
ensure_gh_auth

echo "→ Downloading APK asset(s) from ${OWNER_REPO} ($([[ -n "$TAG_OPT" ]] && echo "tag $TAG_OPT" || echo "latest"))…"
if [ -n "$TAG_OPT" ]; then
  gh release download "$TAG_OPT" --repo "$OWNER_REPO" \
    --pattern "*.apk" --pattern "*.sha256" --clobber --dir "$DEST_DIR"
else
  gh release download --repo "$OWNER_REPO" \
    --pattern "*.apk" --pattern "*.sha256" --clobber --dir "$DEST_DIR"
fi

APK_FILE="$(ls -t "$DEST_DIR"/*.apk 2>/dev/null | head -n1 || true)"
if [ -z "${APK_FILE:-}" ]; then
  echo "ERROR: No *.apk assets found in that release." >&2
  exit 1
fi
echo "✓ APK: $APK_FILE"

# Optional checksum verification (if the release publishes a matching .sha256)
if [ -f "${APK_FILE}.sha256" ]; then
  echo "→ Verifying checksum…"
  (cd "$DEST_DIR" && sha256sum -c "$(basename "${APK_FILE}.sha256")")
else
  echo "↷ No .sha256 file found; skipping checksum verification."
fi

# Optional signature / signer verification
if command -v apksigner >/dev/null 2>&1; then
  echo "→ Verifying APK signature (apksigner)…"
  apksigner verify --print-certs "$APK_FILE" >/tmp/apkcerts.txt
  cat /tmp/apkcerts.txt
  if [ -n "$EXPECTED_FINGERPRINT" ]; then
    FP="$(awk '/SHA-256 digest:/ {print $3}' /tmp/apkcerts.txt | head -n1)"
    norm() { echo "$1" | tr -d ':' | tr '[:lower:]' '[:upper:]'; }
    if [ "$(norm "$FP")" != "$(norm "$EXPECTED_FINGERPRINT")" ]; then
      echo "ERROR: Signer fingerprint mismatch! Expected $EXPECTED_FINGERPRINT but got $FP" >&2
      exit 1
    fi
  fi
else
  echo "↷ apksigner not found; skipping signature verification."
fi

# Install via ADB
echo "→ Installing on connected device (adb)…"
adb devices
adb install -r "$APK_FILE"

echo "🎉 Done."
