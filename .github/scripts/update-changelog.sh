#!/usr/bin/env bash
# Prepend a generated release-notes section to CHANGELOG.md when absent.
# Usage: update-changelog.sh NEW_TAG NOTES_FILE
set -euo pipefail

NEW_TAG="${1:?new tag required (e.g. v2.12.0)}"
NOTES_FILE="${2:?notes file required}"

VERSION="${NEW_TAG#v}"
CHANGELOG="CHANGELOG.md"

if [[ ! -f "$NOTES_FILE" ]]; then
  echo "notes file not found: $NOTES_FILE" >&2
  exit 1
fi

if [[ ! -f "$CHANGELOG" ]]; then
  touch "$CHANGELOG"
fi

if grep -q "^## ${VERSION}$" "$CHANGELOG"; then
  echo "CHANGELOG already contains section for ${VERSION}; skipping update"
  exit 0
fi

section_file="$(mktemp)"
trap 'rm -f "$section_file"' EXIT

# Extract the HashiCorp-style section (## VERSION through end of categorized content).
awk -v version="$VERSION" '
  /^## / {
    if ($0 == "## " version) {
      in_section = 1
      print
      next
    }
    if (in_section) {
      exit
    }
  }
  in_section { print }
' "$NOTES_FILE" > "$section_file"

if [[ ! -s "$section_file" ]]; then
  echo "failed to extract changelog section for ${VERSION} from ${NOTES_FILE}" >&2
  exit 1
fi

existing="$(mktemp)"
trap 'rm -f "$section_file" "$existing"' EXIT
cat "$CHANGELOG" > "$existing"

{
  cat "$section_file"
  if [[ -s "$existing" ]]; then
    printf '\n'
    cat "$existing"
  fi
} > "$CHANGELOG"

echo "Updated CHANGELOG.md with section for ${VERSION}"
