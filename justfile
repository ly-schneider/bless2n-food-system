default:
    @just --list

# Bump VERSION and package.json to a specific semver (e.g., just bump 1.2.3)
[group('release')]
bump version:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! echo "{{version}}" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
      echo "Invalid semver: {{version}}" && exit 1
    fi
    echo "{{version}}" > VERSION
    cd bfs-web-app
    ESCAPED=$(echo "{{version}}" | sed 's/[&/]/\\&/g')
    sed -i '' "s/\"version\": \".*\"/\"version\": \"${ESCAPED}\"/" package.json
    echo "Bumped to {{version}}"
