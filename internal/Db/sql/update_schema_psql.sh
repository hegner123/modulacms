#\!/bin/bash

# Function to update PostgreSQL schema files
update_schema_psql() {
  local folder=$1
  local id_type=$2  # admin_route_id or route_id
  
  echo "Updating $folder schema_psql.sql for $id_type..."
  
  # PostgreSQL schema file
  psql_file="schema/$folder/schema_psql.sql"
  if [ -f "$psql_file" ]; then
    # Remove the column from CREATE TABLE
    sed -i '' -e "s/$id_type INTEGER[^,]*,//" "$psql_file"
    sed -i '' -e "s/$id_type INT[^,]*,//" "$psql_file"
    
    # Remove foreign key constraints
    sed -i '' -e "/CONSTRAINT.*$id_type.*FOREIGN KEY/,/ON DELETE/d" "$psql_file"
    
    echo "Updated $psql_file"
  fi
}

# Admin tables
update_schema_psql "admin_datatype" "admin_route_id"
update_schema_psql "admin_field" "admin_route_id"

# Regular tables
update_schema_psql "datatype" "route_id"
update_schema_psql "field" "route_id"

echo "Schema PostgreSQL update complete\!"
