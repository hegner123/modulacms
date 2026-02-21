#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA_DIR="$SCRIPT_DIR/schema"

generate() {
    local suffix="$1"
    local header="$2"
    local outfile="$SCRIPT_DIR/$3"

    echo "-- $header" > "$outfile"
    echo "-- Order follows sql/schema/ directory numbering" >> "$outfile"
    echo "-- Generated from individual schema files" >> "$outfile"

    for dir in "$SCHEMA_DIR"/*/; do
        local name
        name="$(basename "$dir")"

        # Skip utility and 0_wipe (not part of fresh install schema)
        if [[ "$name" == "utility" || "$name" == "0_wipe" ]]; then
            continue
        fi

        local schema_file="$dir/schema${suffix}.sql"
        if [[ ! -f "$schema_file" ]]; then
            continue
        fi

        echo "" >> "$outfile"
        echo "-- ===== ${name} =====" >> "$outfile"
        echo "" >> "$outfile"
        cat "$schema_file" >> "$outfile"
    done

    echo "Generated $outfile"
}

generate "" "ModulaCMS Schema (SQLite)" "all_schema.sql"
generate "_mysql" "ModulaCMS Schema (MySQL)" "all_schema_mysql.sql"
generate "_psql" "ModulaCMS Schema (PostgreSQL)" "all_schema_psql.sql"
