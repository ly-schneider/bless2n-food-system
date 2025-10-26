#!/usr/bin/env bash
set -euo pipefail

# Fixed repo as requested
OWNER_REPO="ly-schneider/bless2n-food-system"

# Defaults
DEST_DIR="./dist"
TAG_OPT=""            # e.g. v1.2.3; empty = latest
ENV_FILTER=""         # e.g. production|staging; empty = pick newest
APK_NAME=""           # explicit file name to install after download
GH_HOST="${GH_HOST:-github.com}"

# Optional strong check: expected signer fingerprint (SHA-256) for the APK
# Set via env: EXPECTED_FINGERPRINT="AA:BB:...:ZZ"  (colons optional)
EXPECTED_FINGERPRINT="${EXPECTED_FINGERPRINT:-}"

usage() {
  cat <<EOF
Usage: $(basename "$0") [DEST_DIR] [TAG] [--env ENV] [--tag TAG] [--apk FILE]

Examples:
  # Install latest production APK
  $0 ./dist --env production

  # Install staging APK from v1.0.9
  $0 ./dist --tag v1.0.9 --env staging

  # Install a specific asset by name (must exist in release)
  $0 ./dist --tag v1.0.9 --apk bfs-android-app-release-staging.apk

Positional args (kept for backward-compat):
  DEST_DIR  Destination directory for downloads (default: ./dist)
  TAG       Release tag to download from (default: latest)

Flags:
  --env ENV   Filter to APK named bfs-android-app-release-ENV.apk (e.g. production|staging)
  --tag TAG   Same as positional TAG
  --apk FILE  Explicit APK file name to install (after download)
  -h, --help  Show this help
EOF
}

# Parse args (support both positional and flags)
pos_seen=0
while [ $# -gt 0 ]; do
  case "$1" in
    -h|--help)
      usage; exit 0 ;;
    --env)
      ENV_FILTER="${2:-}"; shift 2 ;;
    --tag)
      TAG_OPT="${2:-}"; shift 2 ;;
    --apk)
      APK_NAME="${2:-}"; shift 2 ;;
    --*)
      echo "Unknown flag: $1" >&2; usage; exit 2 ;;
    *)
      # positional
      if [ $pos_seen -eq 0 ]; then DEST_DIR="$1"; pos_seen=1
      elif [ $pos_seen -eq 1 ]; then TAG_OPT="$1"; pos_seen=2
      else echo "Unexpected extra arg: $1" >&2; usage; exit 2; fi
      shift
      continue ;;
  esac
done

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

# Choose APK
APK_FILE=""
if [ -n "$APK_NAME" ]; then
  # Use user-provided explicit file name
  if [ -f "$DEST_DIR/$APK_NAME" ]; then APK_FILE="$DEST_DIR/$APK_NAME"; else
    echo "ERROR: Specified --apk '$APK_NAME' not found in '$DEST_DIR'" >&2; exit 1; fi
elif [ -n "$ENV_FILTER" ]; then
  # Prefer exact naming convention used by the release workflow
  CANDIDATE="$DEST_DIR/bfs-android-app-release-${ENV_FILTER}.apk"
  if [ -f "$CANDIDATE" ]; then
    APK_FILE="$CANDIDATE"
  else
    # Fallback: try any file containing the env in its name
    APK_FILE="$(ls -t "$DEST_DIR"/*"${ENV_FILTER}"*.apk 2>/dev/null | head -n1 || true)"
  fi
else
  # Fallback: newest APK
  APK_FILE="$(ls -t "$DEST_DIR"/*.apk 2>/dev/null | head -n1 || true)"
fi

if [ -z "${APK_FILE:-}" ] || [ ! -f "$APK_FILE" ]; then
  echo "ERROR: No matching *.apk asset found in that release." >&2
  if [ -n "$ENV_FILTER" ]; then echo "       Tried to find env '$ENV_FILTER' (bfs-android-app-release-${ENV_FILTER}.apk)." >&2; fi
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
