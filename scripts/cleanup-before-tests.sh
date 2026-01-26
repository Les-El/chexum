#!/bin/bash
# cleanup-before-tests.sh
# This script should be run BEFORE executing `go test` to ensure /tmp has sufficient space
# It removes accumulated temporary files and Go build artifacts safely

set -e

PATTERNS=(
    "/tmp/go-build*"
    "/tmp/hashi-*"
    "/tmp/checkpoint-*"
    "/tmp/test-*"
)

echo "Cleaning up /tmp to prevent disk space issues during testing..."

# More aggressive cleanup using find with delete option
for pattern in "${PATTERNS[@]}"; do
    # Extract the pattern part after /tmp/
    pat="${pattern##/tmp/}"
    
    # Use find with -delete for safer, more efficient removal
    find /tmp -maxdepth 1 -name "$pat" -type d -exec rm -rf {} + 2>/dev/null || true
    find /tmp -maxdepth 1 -name "$pat" -type f -delete 2>/dev/null || true
done

# Additional cleanup: remove any orphaned Go build directories
if [ -d /tmp ]; then
    find /tmp -maxdepth 1 -type d -name "go-*" -mmin +60 -exec rm -rf {} + 2>/dev/null || true
fi

echo "Cleanup complete. Current /tmp usage:"
df -h /tmp | awk 'NR==2 {print $1 "\t" $5}'

