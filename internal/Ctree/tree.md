# Content Tree Plan
how to build a json object to send to my site

## Tables
1. Route: a definition of a slug
2. Datatype: A definition of a type
3. Fields: A definition of a field that belongs to a type
4. Content Data: An instance of a type assigned to a route
5. Content Field: An instance of a field that belongs to an instance of a type assigned to a route

## Steps

Database Query Strategy:
• Lookup the Route: Query the routes table to find the record by slug.
• Fetch Content Data: Using the route’s ID or slug, fetch the content data instance.
• Get the Datatype Definition: Join or query the datatype table to retrieve the content type information.
• Retrieve Field Definitions: Based on the datatype, fetch all field definitions from the fields table.
• Match Field Values: For each field definition, query the content field table for the corresponding value linked to your content data instance.
