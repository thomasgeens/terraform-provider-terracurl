#!/usr/bin/env bash
# Generate HashiCorp-style release notes from commits since the previous tag.
# Usage: generate-release-notes.sh NEW_TAG [PREV_TAG]
# Output: markdown suitable for CHANGELOG.md and GitHub release notes.
set -euo pipefail

NEW_TAG="${1:?new tag required (e.g. v2.12.0)}"
PREV_TAG="${2:-}"

if [[ -z "$PREV_TAG" ]]; then
  PREV_TAG="$(git describe --tags --abbrev=0 "${NEW_TAG}^" 2>/dev/null || true)"
fi

VERSION="${NEW_TAG#v}"
REPO="${GITHUB_REPOSITORY:-devops-rob/terraform-provider-terracurl}"

if [[ -n "$PREV_TAG" ]]; then
  RANGE="${PREV_TAG}..${NEW_TAG}"
  COMPARE_URL="https://github.com/${REPO}/compare/${PREV_TAG}...${NEW_TAG}"
else
  RANGE="${NEW_TAG}"
  COMPARE_URL=""
fi

should_skip() {
  local subject="$1"
  local lower="${subject,,}"

  if [[ "$lower" =~ ^merge[[:space:]] ]]; then
    return 0
  fi
  if [[ "$lower" =~ ^chore: ]]; then
    return 0
  fi
  if [[ "$lower" =~ fix[[:space:]]gofmt ]]; then
    return 0
  fi
  if [[ "$lower" =~ fix[[:space:]]linter ]]; then
    return 0
  fi
  if [[ "$lower" =~ fix[[:space:]]ci[[:space:]]linter ]]; then
    return 0
  fi
  if [[ "$lower" =~ ^align[[:space:]] ]]; then
    return 0
  fi

  return 1
}

commit_group() {
  local subject="$1"
  local lower="${subject,,}"

  if should_skip "$subject"; then
    echo "skip"
    return
  fi

  if [[ "$lower" =~ ^add[[:space:]] ]] || \
     [[ "$lower" =~ ^feat: ]] || \
     [[ "$lower" =~ ^feat[[:space:]] ]] || \
     [[ "$lower" =~ ^enhancement: ]] || \
     [[ "$subject" == *"(Closes #"* ]] || \
     [[ "$subject" == *"(closes #"* ]]; then
    echo "enhancements"
    return
  fi

  if [[ "$lower" =~ ^fix[[:space:]] ]] || \
     [[ "$lower" =~ ^fix: ]] || \
     [[ "$lower" =~ ^bug: ]] || \
     [[ "$lower" =~ ^patch: ]]; then
    echo "bug_fixes"
    return
  fi

  echo "other"
}

enhancement_lines=()
bug_fix_lines=()
other_lines=()

while IFS= read -r subject; do
  [[ -z "$subject" ]] && continue
  case "$(commit_group "$subject")" in
    enhancements) enhancement_lines+=("$subject") ;;
    bug_fixes) bug_fix_lines+=("$subject") ;;
    other) other_lines+=("$subject") ;;
  esac
done < <(git log --format=%s "$RANGE" 2>/dev/null || true)

print_section() {
  local title="$1"
  shift
  if [[ $# -eq 0 ]]; then
    return
  fi
  printf '%s:\n\n' "$title"
  for line in "$@"; do
    printf -- '- %s\n' "$line"
  done
  printf '\n'
}

printf '# %s\n\n' "$NEW_TAG"
printf 'Terraform provider for HTTP requests via Terraform. Install from the [Terraform Registry](https://registry.terraform.io/providers/devops-rob/terracurl/latest).\n\n'

if [[ -n "$COMPARE_URL" ]]; then
  printf '**Full changelog:** [%s...%s](%s)\n\n' "$PREV_TAG" "$NEW_TAG" "$COMPARE_URL"
fi

printf '## %s\n\n' "$VERSION"

print_section "ENHANCEMENTS" "${enhancement_lines[@]}"
print_section "BUG FIXES" "${bug_fix_lines[@]}"
print_section "OTHER" "${other_lines[@]}"

if [[ ${#enhancement_lines[@]} -eq 0 && ${#bug_fix_lines[@]} -eq 0 && ${#other_lines[@]} -eq 0 ]]; then
  printf '_No categorized commits in range %s._\n' "$RANGE"
fi
