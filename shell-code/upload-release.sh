#!/bin/bash

# Variables
VERSION_FILE="VERSION.txt"

GITHUB_TOKEN="${GITHUB_TOKEN}"  # Replace with your GitHub token

REPO="Direct-Dev-Ru/binaries.git"  # Replace with your GitHub username/repo

TAG=go-ansible-vault.$(cat "$VERSION_FILE")

echo $TAG

RELEASE_NAME="Binaries ${TAG}"  # Replace with your release title

RELEASE_DIR="/home/su/projects/golang/ansible-vault/binaries-for-upload"  

# Create a new release
# response=$(curl -s -X POST \
#     -H "Authorization: Bearer ${GITHUB_TOKEN}" \
#     -H "Accept: application/vnd.github+json" \
#     -H "X-GitHub-Api-Version: 2022-11-28" \
#     https://api.github.com/repos/$REPO/releases \
#     -d "{\"tag_name\": \"$TAG\", \"name\": \"$RELEASE_NAME\"}")

body="{\"tag_name\":\"${TAG}\", \"target_commitish\":\"main\", \"name\":\"${TAG}\", \
  \"body\":\"${TAG}\", \"draft\":false, \"prerelease\":false, \"generate_release_notes\":false}"

echo $body

response=$(curl -L -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/direct-dev-ru/binaries/releases \
  -d $body)


echo $response

# Extract the upload URL from the response
upload_url=$(echo "$response" | jq -r '.upload_url' | sed "s/{?name,label}//")

# Check if the release was created successfully
if [[ "$response" == *"Not Found"* ]]; then
    echo "Error: Repository not found or invalid token."
    exit 1
fi

# Upload each binary file
for file in "$RELEASE_DIR"/*; do
    if [[ -f "$file" ]]; then
        filename=$(basename "$file")
        echo "Uploading $filename..."
        response=$(curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
            -H "Content-Type: application/octet-stream" \
            "$upload_url?name=$filename" \
            --data-binary @"$file")
        echo $response    
    fi
done

echo "All binaries uploaded successfully."
