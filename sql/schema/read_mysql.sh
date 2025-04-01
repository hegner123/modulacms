#!/usr/bin/env bash
# Define the output file in the parent directory
OUTPUT="../../internal/Db/sql/setup_mysql.sql"

# Optional: Clear the output file if it exists, or create it if it doesn't
> "$OUTPUT"

# Find all schema_mysql.sql files in subdirectories and append their contents
find . -maxdepth 2 -type f -name schema_mysql.sql -print0 | sort -z -V | while IFS= read -r -d '' file; do
  cat "$file" >> $OUTPUT
done

echo "Contents of all schema_mysql.sql files have been appended to $OUTPUT"

