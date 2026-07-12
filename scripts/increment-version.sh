#!/bin/env bash

# Increments the VERSION file. Format: YYYY.MM.XXX where:
#   - YYYY is the current year
#   - MM is the current month (01-12)
#   - XXX is the build number starting from 001
# If the year or month has changed, the build number resets to 001.

VERSION_FILE="${1:-VERSION}"

if [ ! -f "$VERSION_FILE" ]; then
    echo "Error: VERSION file not found."
    exit 1
fi

CURRENT_VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
echo "Current version: $CURRENT_VERSION"

CURRENT_YEAR=$(date +%Y)
CURRENT_MONTH=$(date +%m)

IFS='.' read -r VERSION_YEAR VERSION_MONTH VERSION_BUILD <<< "$CURRENT_VERSION"

if [ "$VERSION_YEAR" != "$CURRENT_YEAR" ] || [ "$VERSION_MONTH" != "$CURRENT_MONTH" ]; then
    NEW_BUILD="001"
else
    BUILD_NUM=$((10#$VERSION_BUILD + 1))
    NEW_BUILD=$(printf "%03d" $BUILD_NUM)
fi

NEW_VERSION="${CURRENT_YEAR}.${CURRENT_MONTH}.${NEW_BUILD}"

echo "$NEW_VERSION" > "$VERSION_FILE"

echo "New version: $NEW_VERSION"
