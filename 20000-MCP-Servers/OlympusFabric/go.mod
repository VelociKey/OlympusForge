module github.com/VelociKey/OlympusFabric/20000-MCP-Servers/OlympusFabric

go 1.25.7

require (
	github.com/VelociKey/Olympus2/pkg/whisper v0.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/VelociKey/Olympus2/pkg/actor v0.0.0-00010101000000-000000000000 // indirect

replace github.com/VelociKey/Olympus2/pkg/whisper => c:/aAntigravitySpace/Olympus2/90000-Enablement/pkg/whisper

replace github.com/VelociKey/Olympus2/pkg/actor => c:/aAntigravitySpace/Olympus2/90000-Enablement/pkg/actor
