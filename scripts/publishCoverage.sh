#!/bin/bash
#******************************************************************************
# Copyright 2022 IBM Corp.
# Licensed under the Apache License, Version  2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#******************************************************************************
set -euo pipefail

echo "===== Publishing the coverage results ====="

WORKDIR="$GITHUB_WORKSPACE/gh-pages"
NEW_COVERAGE_SOURCE="$GITHUB_WORKSPACE/cover.html"
BADGE_COLOR="red"
GREEN_THRESHOLD=85
YELLOW_THRESHOLD=50

# Helper: extract coverage % from cover.html
get_coverage() {
    local file="$1"
    if [[ -f "$file" ]]; then
        grep "%)" "$file" \
          | sed 's/[][()><%]/ /g' \
          | awk '{s+=$4}END{if(NR>0)print s/NR; else print 0}'
    else
        echo "0"
    fi
}

# Base branch for comparison
if [[ "$GITHUB_EVENT_NAME" == "pull_request" ]]; then
    BASE_BRANCH="$GITHUB_BASE_REF"
else
    BASE_BRANCH="$GITHUB_REF_NAME"
fi

# Calculate new coverage
NEW_COVERAGE=$(get_coverage "$NEW_COVERAGE_SOURCE")
NEW_COVERAGE=$(printf "%.2f" "$NEW_COVERAGE")

# Clone gh-pages
mkdir -p "$WORKDIR"
cd "$WORKDIR"

if ! git clone -q -b gh-pages "https://x-access-token:$GHE_TOKEN@github.com/$GITHUB_REPOSITORY.git" . 2>/dev/null; then
    echo "gh-pages branch not found → creating it"
    git init -q
    git checkout -b gh-pages
fi

git config user.name "github-actions[bot]"
git config user.email "github-actions[bot]@users.noreply.github.com"

# Calculate old coverage (from base branch)
COVERAGE_DIR="coverage/$BASE_BRANCH"
OLD_COVER_HTML="$COVERAGE_DIR/cover.html"
OLD_COVERAGE=$(get_coverage "$OLD_COVER_HTML")
OLD_COVERAGE=$(printf "%.2f" "$OLD_COVERAGE")

echo "===== Coverage comparison ====="
echo "Old Coverage: $OLD_COVERAGE%"
echo "New Coverage: $NEW_COVERAGE%"

# Update reports
mkdir -p "$COVERAGE_DIR"
mkdir -p "coverage/$GITHUB_SHA"
cp "$NEW_COVERAGE_SOURCE" "$COVERAGE_DIR/cover.html"
cp "$NEW_COVERAGE_SOURCE" "coverage/$GITHUB_SHA/cover.html"

# Badge color
if (( $(echo "$NEW_COVERAGE > $GREEN_THRESHOLD" | bc -l) )); then
    BADGE_COLOR="green"
elif (( $(echo "$NEW_COVERAGE > $YELLOW_THRESHOLD" | bc -l) )); then
    BADGE_COLOR="yellow"
fi

curl -s "https://img.shields.io/badge/coverage-${NEW_COVERAGE}%25-${BADGE_COLOR}.svg" \
     > "$COVERAGE_DIR/badge.svg"

# Result message
if (( $(echo "$OLD_COVERAGE > $NEW_COVERAGE" | bc -l) )); then
    RESULT_MESSAGE="Coverage decreased from **$OLD_COVERAGE%** → **$NEW_COVERAGE%**"
elif (( $(echo "$OLD_COVERAGE == $NEW_COVERAGE" | bc -l) )); then
    RESULT_MESSAGE="Coverage remained the same at **$NEW_COVERAGE%**"
else
    RESULT_MESSAGE="Coverage increased from **$OLD_COVERAGE%** → **$NEW_COVERAGE%**"
fi

# Push to gh-pages (only on push events) or comment on PR
if [[ "$GITHUB_EVENT_NAME" == "push" ]] || [[ "$GITHUB_EVENT_NAME" == "workflow_dispatch" ]]; then
    git add .
    git commit -m "Coverage: $GITHUB_SHA (run $GITHUB_RUN_NUMBER)" || echo "Nothing to commit"
    git push "https://x-access-token:$GHE_TOKEN@github.com/$GITHUB_REPOSITORY.git" gh-pages
fi

if [[ "$GITHUB_EVENT_NAME" == "pull_request" ]]; then
    PR_NUMBER=$(jq -r .pull_request.number "$GITHUB_EVENT_PATH")
    curl -s -X POST \
      -H "Authorization: token $GHE_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"body\": \"$RESULT_MESSAGE\"}" \
      "https://api.github.com/repos/$GITHUB_REPOSITORY/issues/$PR_NUMBER/comments"
fi

echo "===== Coverage publishing finished ====="