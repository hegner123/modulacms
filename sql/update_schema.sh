#\!/bin/bash

# Function to update MySQL and PostgreSQL files
update_files() {
  local folder=$1
  local id_type=$2  # admin_route_id or route_id
  
  echo "Updating $folder files for $id_type..."
  
  # MySQL queries file
  mysql_file="schema/$folder/queries_mysql.sql"
  if [ -f "$mysql_file" ]; then
    # Remove the column from CREATE TABLE
    sed -i '' -e "s/$id_type INT[^,]*,//" "$mysql_file"
    
    # Remove foreign key constraints
    sed -i '' -e "/CONSTRAINT.*$id_type.*FOREIGN KEY/,/ON DELETE/d" "$mysql_file"
    
    # Remove from INSERT statements
    sed -i '' -e "s/$id_type,//" "$mysql_file"
    
    # Remove from UPDATE statements
    sed -i '' -e "s/SET $id_type = ?,/SET /" "$mysql_file"
    sed -i '' -e "s/$id_type = ?,//" "$mysql_file"
    
    # Remove ListByRouteId queries
    sed -i '' -e "/-- name: List.*ByRouteId/,/WHERE $id_type = ?;/d" "$mysql_file"
    
    echo "Updated $mysql_file"
  fi
  
  # PostgreSQL queries file
  psql_file="schema/$folder/queries_psql.sql"
  if [ -f "$psql_file" ]; then
    # Remove the column from CREATE TABLE
    sed -i '' -e "s/$id_type INT[^,]*,//" "$psql_file"
    sed -i '' -e "s/$id_type INTEGER[^,]*,//" "$psql_file"
    
    # Remove foreign key constraints
    sed -i '' -e "/CONSTRAINT.*$id_type.*FOREIGN KEY/,/ON DELETE/d" "$psql_file"
    
    # Remove from INSERT statements
    sed -i '' -e "s/$id_type,//" "$psql_file"
    
    # Remove from UPDATE statements
    sed -i '' -e "s/SET $id_type = \$[0-9],/SET /" "$psql_file"
    sed -i '' -e "s/$id_type = \$[0-9],//" "$psql_file"
    
    # Remove ListByRouteId queries
    sed -i '' -e "/-- name: List.*ByRouteId/,/WHERE $id_type = \$[0-9];/d" "$psql_file"
    sed -i '' -e "/-- name: GetRootAdminDtByAdminRtId/,/ORDER BY.*_id;/d" "$psql_file"
    
    echo "Updated $psql_file"
  fi
}

# Admin tables
update_files "admin_field" "admin_route_id"

# Regular tables
update_files "datatype" "route_id"
update_files "field" "route_id"

echo "Schema update complete\!"
