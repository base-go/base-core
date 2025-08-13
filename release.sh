#!/bin/bash

set -e

# Check if version is provided
if [ -z "$1" ]; then
    echo "Usage: ./release.sh <version>"
    echo "Example: ./release.sh 2.0.0"
    exit 1
fi

# Handle version with or without 'v' prefix
if [[ "$1" == v* ]]; then
    VERSION="$1"
else
    VERSION="v$1"
fi

echo "Creating Base Framework release $VERSION..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Error: Not in a git repository"
    exit 1
fi

# Check if there are uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "Error: There are uncommitted changes. Please commit or stash them first."
    exit 1
fi

# Create clean release directory
echo "Creating clean release directory..."
RELEASE_DIR="base-framework-$VERSION"
mkdir -p "$RELEASE_DIR"

# Copy files excluding development files
rsync -av --exclude='.git' \
    --exclude='CHANGELOG.md' \
    --exclude='release.sh' \
    --exclude='docs.md' \
    --exclude='*.test' \
    --exclude='.DS_Store' \
    --exclude='basecmd' \
    --exclude='.claude' \
    --exclude='packages' \
    ./ "$RELEASE_DIR/"

# Remove binary executables but keep source files
rm -f "$RELEASE_DIR/base"

# Create clean archive
echo "Creating clean archive..."
tar -czf "base-framework-$VERSION.tar.gz" "$RELEASE_DIR"
zip -r "base-framework-$VERSION.zip" "$RELEASE_DIR"

# Clean up
rm -rf "$RELEASE_DIR"

# Create and push tag
echo "Creating git tag $VERSION..."
git tag -a "$VERSION" -m "Base Framework $VERSION

$(git log --pretty=format:'- %s' $(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")..HEAD | head -20)"

echo "Pushing tag to remote..."
git push origin "$VERSION"

# Create GitHub release (if gh CLI is available)
if command -v gh &> /dev/null; then
    echo "Creating GitHub release..."
    
    # Extract changelog for this version from CHANGELOG.md
    CHANGELOG_SECTION=""
    if [ -f "CHANGELOG.md" ]; then
        # Extract the section for this version from CHANGELOG.md
        CHANGELOG_SECTION=$(awk "/^## \[${VERSION#v}\]/ {flag=1; next} /^## \[/ && flag {flag=0} flag" CHANGELOG.md)
    fi
    
    if [ -n "$CHANGELOG_SECTION" ]; then
        RELEASE_NOTES="Base Framework $VERSION

$CHANGELOG_SECTION

## Installation

To use this version of Base Framework:

\`\`\`bash
# Update your existing project
base update

# Or create a new project (will use latest framework)
base new myproject
\`\`\`

## CLI Compatibility

This framework version is compatible with Base CLI v$VERSION and later."
    else
        RELEASE_NOTES="Base Framework $VERSION

What's new:
$(git log --pretty=format:'- %s' $(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")..HEAD | head -10)

## Installation

To use this version of Base Framework:

\`\`\`bash
# Update your existing project
base update

# Or create a new project (will use latest framework)
base new myproject
\`\`\`

## CLI Compatibility

This framework version is compatible with Base CLI v$VERSION and later."
    fi

    gh release create "$VERSION" \
        --title "Base Framework $VERSION" \
        --notes "$RELEASE_NOTES" \
        "base-framework-$VERSION.tar.gz" \
        "base-framework-$VERSION.zip"
    
    # Clean up archive files
    rm -f "base-framework-$VERSION.tar.gz" "base-framework-$VERSION.zip"
        
    echo "✅ GitHub release created successfully!"
else
    echo "⚠️  GitHub CLI (gh) not found. Please create the release manually at:"
    echo "   https://github.com/base-go/base/releases/new?tag=$VERSION"
fi

echo ""
echo "🎉 Base Framework release $VERSION completed successfully!"
echo ""
echo "Next steps:"
echo "1. Update CLI repository to reference this framework version"
echo "2. Test the 'base update' command with existing projects"  
echo "3. Create new projects to verify they use the updated framework"
echo ""