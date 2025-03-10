#!/usr/bin/env bash
# Define the output file in the parent directory
OUTPUT="../all_schema_mysql.sql"

# Optional: Clear the output file if it exists, or create it if it doesn't
> "$OUTPUT"

# Find all schema_mysql.sql files in subdirectories and append their contents
find . -type f -name "schema_mysql.sql" -exec cat {} \; >> "$OUTPUT"

echo "Contents of all schema_mysql.sql files have been appended to $OUTPUT"

