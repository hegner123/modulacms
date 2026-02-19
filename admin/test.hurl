GET http://localhost:8080/api/v1/plugins/task_tracker/categories
Authorization: Bearer Ty4J9FDEfoSIysH57uKeiznXxfNrFVVO8cJyjBQbnuc

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$" isCollection
