#!/usr/bin/env bash

# Check if exactly four arguments are provided.
if [ "$#" -ne 4 ]; then
    echo "Usage: $0 <username> <password> <database> <output_file>"
    exit 1
fi

# Assign command line arguments to variables.
USERNAME="$1"
PASSWORD="$2"
DATABASE="$3"
OUTPUT_FILE="$4"

# Set the PGPASSWORD environment variable for pg_dump.
export PGPASSWORD="$PASSWORD"

# Execute pg_dump using the provided arguments.
pg_dump -U "$USERNAME" -d "$DATABASE" -f "$OUTPUT_FILE"

# Check if the pg_dump command succeeded.
if [ $? -eq 0 ]; then
    echo "Database '$DATABASE' backup successful. Output file: '$OUTPUT_FILE'."
else
    echo "Error: Backup of database '$DATABASE' failed."
    exit 1
fi
