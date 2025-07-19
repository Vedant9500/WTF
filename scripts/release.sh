#!/bin/bash

# WTF Release Script
# Usage: ./scripts/release.sh v1.1.0

set -e

VERSION=${1:-}

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.1.0"
    exit 1
fi

echo "🚀 Preparing release $VERSION"

# Verify we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "❌ Must be on main branch. Currently on: $CURRENT_BRANCH"
    exit 1
fi

# Verify working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ Working directory is not clean. Please commit or stash changes."
    git status --short
    exit 1
fi

# Update version in version.go
echo "📝 Updating version in internal/version/version.go"
sed -i.bak "s/Version = \".*\"/Version = \"${VERSION#v}\"/" internal/version/version.go
rm internal/version/version.go.bak

# Update README badge
echo "📝 Updating README.md version badge"
sed -i.bak "s/version-.*-blue/version-${VERSION#v}-blue/" README.md
rm README.md.bak

# Update DONE.md status
echo "📝 Updating DONE.md status"
sed -i.bak "s/Status\*\*: v.* Production/Status**: ${VERSION} Production/" DONE.md
rm DONE.md.bak

# Run tests
echo "🧪 Running tests"
go test ./...

# Build to verify everything works
echo "🔨 Building to verify"
go build -o build/wtf ./cmd/wtf

# Test the version
echo "✅ Testing version output"
./build/wtf --version

# Commit version updates
echo "📝 Committing version updates"
git add internal/version/version.go README.md DONE.md
git commit -m "chore: bump version to $VERSION"

# Create and push tag
echo "🏷️  Creating and pushing tag $VERSION"
git tag -a "$VERSION" -m "Release $VERSION"
git push origin main
git push origin "$VERSION"

echo "🎉 Release $VERSION has been triggered!"
echo "📋 Check the Actions tab in GitHub to monitor the build progress"
echo "🔗 https://github.com/Vedant9500/WTF/actions"

# Clean up
rm -f build/wtf
