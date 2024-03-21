#!/bin/bash

# Check if the new port number is provided as an argument
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 NEW_PORT"
    exit 1
fi

# Assign the first script argument to NEW_PORT
NEW_PORT="$1"

# Base directory to start from, "." means the current directory
BASE_DIRECTORY="."

# Find all .yaml files recursively starting from the base directory
# and perform an in-place search and replace from 8443 to the new port
find "$BASE_DIRECTORY" -type f -name "*.yaml" -print0 | xargs -0 sed -i '' -e "s/8443/${NEW_PORT}/g"
echo "Replacement complete. All occurrences of 8443 have been changed to ${NEW_PORT}."

