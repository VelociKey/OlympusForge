module olympus.fleet/00SDLC/OlympusForge

go 1.26.0

require (
	connectrpc.com/connect v1.19.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace connectrpc.com/connect => ./81200-Logic-Libraries/connectrpc/connect-go-1.18.1

// mcp-go replacement removed (not found in workspace)
