#!/usr/bin/env bash
# Script to verify gomod2nix.toml is up to date with go.mod
# Usage: ./scripts/check-gomod2nix.sh

set -e

echo "Checking if gomod2nix.toml is up to date..."

# Create a temporary copy of current gomod2nix.toml
if [ -f gomod2nix.toml ]; then
    cp gomod2nix.toml /tmp/gomod2nix.toml.backup
fi

# Generate new gomod2nix.toml
gomod2nix generate

# Compare with backup
if [ -f /tmp/gomod2nix.toml.backup ]; then
    if ! diff -q /tmp/gomod2nix.toml.backup gomod2nix.toml > /dev/null 2>&1; then
        echo "❌ gomod2nix.toml is out of date!"
        echo ""
        echo "Please run: gomod2nix generate"
        echo "Or commit your changes to run lefthook pre-commit hook"
        exit 1
    fi
fi

echo "✅ gomod2nix.toml is up to date!"
