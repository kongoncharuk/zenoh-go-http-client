#!/usr/bin/env bash
set -euo pipefail

VERSION_FILE="VERSION"

# 1. Ensure version file exists
if [[ ! -f "$VERSION_FILE" ]]; then
  echo "❌ Error: $VERSION_FILE not found."
  exit 1
fi

# 2. Read current version
CURRENT_VERSION=$(tr -d '[:space:]' < "$VERSION_FILE")

# 3. Parse version (semver: major.minor.patch)
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION" || {
  echo "❌ Error: VERSION file must contain a semantic version like 1.2.3"
  exit 1
}

# 4. Increment patch
PATCH=$((PATCH + 1))
NEW_VERSION="$MAJOR.$MINOR.$PATCH"

# 5. Save new version back to file
echo "$NEW_VERSION" > "$VERSION_FILE"

# 6. Commit version bump
git add "$VERSION_FILE"
git commit -m "Bump version to v$NEW_VERSION"

# 7. Create annotated tag
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"

# 8. Push commit and tag
git push origin "v$NEW_VERSION"

echo "✅ Released v$NEW_VERSION"