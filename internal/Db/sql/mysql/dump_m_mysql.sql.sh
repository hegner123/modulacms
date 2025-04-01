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

# Execute mysqldump using the provided arguments.
mysqldump -u "$USERNAME" -p"$PASSWORD" "$DATABASE" > "$OUTPUT_FILE"

# Check if the mysqldump command succeeded.
if [ $? -eq 0 ]; then
    echo "Database '$DATABASE' backup successful. Output file: '$OUTPUT_FILE'."
else
    echo "Error: Backup of database '$DATABASE' failed."
    exit 1
fi

