#!/usr/bin/env bash

# Check if exactly two arguments are provided.
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <database_file> <output_file>"
    exit 1
fi

# Assign command line arguments to variables.
DATABASE="$1"
OUTPUT_FILE="$2"

# Run sqlite3 dump command and redirect output to the specified file.
sqlite3 "$DATABASE" ".dump" > "$OUTPUT_FILE"

# Check if the sqlite3 command succeeded.
if [ $? -eq 0 ]; then
    echo "Database '$DATABASE' dump successful. Output file: '$OUTPUT_FILE'."
else
    echo "Error: Dump of database '$DATABASE' failed."
    exit 1
fi

