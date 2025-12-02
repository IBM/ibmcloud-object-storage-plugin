#!/bin/bash
#******************************************************************************
# Copyright 2021 IBM Corp.
# Licensed under the Apache License, Version 2.0 (the "License");
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
set -e

echo "Publishing coverage results..."

# Skip if no token (safe for forks)
if [ -z "$GHE_TOKEN" ]; then
  echo "GHE_TOKEN not set → skipping coverage publish"
  exit 0
fi

REPO_SLUG="${GITHUB_REPOSITORY}"
BRANCH="${GITHUB_REF_NAME:-master}"
COMMIT="${GITHUB_SHA}"
RUN_ID="${GITHUB_RUN_ID}"

# Temporary directory for gh-pages
TEMP_DIR=$(mktemp -d)
echo "Cloning gh-pages branch into $TEMP_DIR"

# Try to clone gh-pages, create if it doesn't exist
if ! git clone -q --depth 1 -b gh-pages "https://x-access-token:$GHE_TOKEN@github.com/$REPO_SLUG.git" "$TEMP_DIR" 2>/dev/null; then
  echo "gh-pages branch not found → creating new one"
  git init "$TEMP_DIR"
  cd "$TEMP_DIR"
  git checkout -b gh-pages
else
  cd "$TEMP_DIR"
fi

# Git config
git config user.name "github-actions[bot]"
git config user.email "github-actions[bot]@users.noreply.github.com"

# Create directories
mkdir -p "coverage/$BRANCH"
mkdir -p "coverage/$COMMIT"

# Copy coverage report
cp "$GITHUB_WORKSPACE/cover.html" "coverage/$BRANCH/cover.html"
cp "$GITHUB_WORKSPACE/cover.html" "coverage/$COMMIT/cover.html"

# Calculate coverage percentage
NEW_COVERAGE=$(grep -o '[0-9.]\+%' "coverage/$BRANCH/cover.html" | sed 's/%//' | awk '{sum += $1; count++} END {printf "%.4f", sum/count}')

echo "Current coverage: ${NEW_COVERAGE}%"

# Determine badge color
BADGE_COLOR="red"
if (( $(echo "$NEW_COVERAGE >= 85" | bc -l) )); then
  BADGE_COLOR="brightgreen"
elif (( $(echo "$NEW_COVERAGE >= 50" | bc -l) )); then
  BADGE_COLOR="yellow"
fi

# Generate badge
curl -s "https://img.shields.io/badge/coverage-${NEW_COVERAGE}%25-${BADGE_COLOR}.svg" -o "coverage/$BRANCH/badge.svg"

# Commit and push (normal push — NO --force)
git add .
git commit -m "Coverage update: $COMMIT (run $RUN_ID)" || echo "No changes to commit"
git push "https://x-access-token:$GHE_TOKEN@github.com/$REPO_SLUG.git" gh-pages

echo "Coverage published!"
echo "Badge URL: https://$REPO_SLUG.github.io/coverage/$BRANCH/badge.svg"

# Post comment on PR if applicable
if [ "$GITHUB_EVENT_NAME" = "pull_request" ]; then
  PR_NUMBER=$(echo "$GITHUB_REF" | awk -F / '{print $3}')
  MESSAGE="**Code Coverage:** ${NEW_COVERAGE}%  
![coverage](https://$REPO_SLUG.github.io/coverage/$BRANCH/badge.svg)"

  curl -s -X POST \
    -H "Authorization: token $GHE_TOKEN" \
    -H "Content-Type: application/json" \
    "https://api.github.com/repos/$REPO_SLUG/issues/$PR_NUMBER/comments" \
    -d "{\"body\": \"$MESSAGE\"}"
fi