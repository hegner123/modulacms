module github.com/hegner123/modulacms/mcp

go 1.24

require (
	github.com/hegner123/modulacms/sdks/go v0.0.0
	github.com/mark3labs/mcp-go v0.32.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
)

replace github.com/hegner123/modulacms/sdks/go => ../sdks/go
