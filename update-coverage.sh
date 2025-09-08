#!/bin/bash

# Generate test coverage and update README badge

echo "🧪 Running tests with coverage..."

# Run tests and generate coverage
go test -coverprofile=coverage.out ./... -coverpkg=./... > /dev/null 2>&1

# Extract coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}')

echo "📊 Current coverage: $COVERAGE"

# Update README badge
if [[ "$COVERAGE" =~ ^[0-9]+\.[0-9]+% ]]; then
    # Remove the % for comparison
    COVERAGE_NUM=${COVERAGE%\%}
    
    # Determine badge color based on coverage
    if (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
        COLOR="brightgreen"
    elif (( $(echo "$COVERAGE_NUM >= 60" | bc -l) )); then
        COLOR="green"
    elif (( $(echo "$COVERAGE_NUM >= 40" | bc -l) )); then
        COLOR="yellow"
    elif (( $(echo "$COVERAGE_NUM >= 20" | bc -l) )); then
        COLOR="orange"
    else
        COLOR="red"
    fi
    
    # Escape the % for sed
    COVERAGE_ESCAPED=${COVERAGE/\%/\\%}
    
    # Update the badge in README
    sed -i "s/Coverage-[0-9.]*%25-[a-z]*/Coverage-${COVERAGE_ESCAPED}25-${COLOR}/" README.md
    
    echo "✅ README badge updated to ${COVERAGE} (${COLOR})"
else
    echo "❌ Could not extract coverage percentage"
    exit 1
fi

# Clean up
rm -f coverage.out

echo "🎉 Coverage update complete!"
