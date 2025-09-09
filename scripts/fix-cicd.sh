#!/bin/bash

# Script to fix CI/CD workflow issues
# 1. Fix Docker repository name case issue
# 2. Fix API tests server startup issue

echo "🔧 Fixing CI/CD workflow issues..."

# Backup the original file
cp .github/workflows/cicd.yml .github/workflows/cicd.yml.backup

# Apply fixes using sed to avoid YAML corruption
echo "📝 Applying repository name lowercase fixes..."

# Fix the versioning section - add repo name conversion
sed -i '/imageName="\${{ env\.REGISTRY }}\/\${{ github\.repository }}:$version"/c\
          # Convert repository name to lowercase for Docker registry compatibility\
          repo_name=$(echo "${{ github.repository }}" | tr '\''[:upper:]'\'' '\''[:lower:]'\'')\
          imageName="${{ env.REGISTRY }}/$repo_name:$version"' .github/workflows/cicd.yml

sed -i '/imageTag="\${{ env\.REGISTRY }}\/\${{ github\.repository }}:$tag"/c\
          imageTag="${{ env.REGISTRY }}/$repo_name:$tag"' .github/workflows/cicd.yml

# Add repository name preparation step before metadata extraction
sed -i '/- name: Extract metadata/i\
      - name: Prepare repository name\
        id: repo\
        run: |\
          repo_name=$(echo "${{ github.repository }}" | tr '\''[:upper:]'\'' '\''[:lower:]'\'')\
          echo "name=$repo_name" >> $GITHUB_OUTPUT\
' .github/workflows/cicd.yml

# Fix the images line in metadata extraction
sed -i 's|images: \${{ env\.REGISTRY }}/\${{ github\.repository }}|images: ${{ env.REGISTRY }}/${{ steps.repo.outputs.name }}|' .github/workflows/cicd.yml

echo "✅ Repository name fixes applied"

echo "🐳 Checking Docker service configuration for API tests..."

# The API tests should already be using Docker services, but let's verify the health check
# Add a more robust health check for the API
sed -i '/Wait for API to be ready/,/done/c\
      - name: Wait for API to be ready\
        run: |\
          echo "## 🌐 Waiting for API to be ready..." >> $GITHUB_STEP_SUMMARY\
          for i in {1..60}; do\
            if curl -f -s http://localhost:8080/ > /dev/null 2>&1; then\
              echo "✅ API is ready after $i attempts!"\
              echo "- **API Status:** ✅ Ready" >> $GITHUB_STEP_SUMMARY\
              break\
            fi\
            if [ $i -eq 60 ]; then\
              echo "❌ API failed to start after 60 attempts"\
              echo "- **API Status:** ❌ Failed to start" >> $GITHUB_STEP_SUMMARY\
              exit 1\
            fi\
            echo "⏳ Waiting for API... (attempt $i/60)"\
            sleep 2\
          done' .github/workflows/cicd.yml

echo "✅ API health check improved"

echo "🎯 All fixes applied successfully!"
echo "📋 Summary of changes:"
echo "  - ✅ Repository name converted to lowercase"
echo "  - ✅ Docker metadata uses lowercase name"
echo "  - ✅ API health check timeout increased to 2 minutes"
echo "  - ✅ Better error handling for API startup"
