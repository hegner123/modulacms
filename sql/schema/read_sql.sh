#!/usr/bin/env bash
# Define the output file in the parent directory
OUTPUT="../../all_schema.sql"

# Optional: Clear the output file if it exists, or create it if it doesn't
> "$OUTPUT"

# Find all schema.sql files in subdirectories and append their contents
find . -type f -name "schema.sql" -exec cat {} \; >> "$OUTPUT"

echo "Contents of all schema.sql files have been appended to $OUTPUT"

