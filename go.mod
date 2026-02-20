module OlympusForge

go 1.25.7

// Local Resolution
replace OlympusForge/ZC0400-Sovereign-Source/mcp-go => ./ZC0400-Sovereign-Source/mcp-go
replace aihub-forge => ./90000-Enablement-Labs/900-Forge
replace github.com/mark3labs/mcp-go => ./ZC0400-Sovereign-Source/mcp-go

require (
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
