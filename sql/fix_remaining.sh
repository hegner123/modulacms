#\!/bin/bash

# Manually fix the remaining issues
for file in $(grep -l -r -E "admin_route_id|route_id" schema/admin_datatype/ schema/admin_field/ schema/datatype/ schema/field/ | grep -v "_content_" | grep -v "__pycache__"); do
  echo "Fixing $file"
  
  # First remove any lines with admin_route_id or route_id
  sed -i '' -e '/admin_route_id/d' -e '/route_id/d' "$file"
  
  echo "Fixed $file"
done

echo "All remaining issues fixed\!"
