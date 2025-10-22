#!/bin/bash

execute_command() {
    curl -s -X POST "http://localhost:8085/api/execute" \
        -H "Content-Type: application/json" \
        -d "{\"prompt\": \"$1\", \"verbose\": \"$2\"}" | \
        jq -r '.'
}

execute_command "$1" "$2"