module OlympusForge

go 1.25.7

// Local Resolution
replace Olympus2 => ../Olympus2

replace OlympusActors-Cognition => ../OlympusActors-Cognition

replace OlympusActors-Delegation => ../OlympusActors-Delegation

replace OlympusAscent => ../OlympusAscent

replace OlympusAssurance => ../OlympusAssurance

replace OlympusAtelier => ../OlympusAtelier

replace OlympusFabric => ../OlympusFabric

replace OlympusGCP-Compute => ../OlympusGCP-Compute

replace OlympusGCP-Data => ../OlympusGCP-Data

replace OlympusGCP-Events => ../OlympusGCP-Events

replace OlympusGCP-FinOps => ../OlympusGCP-FinOps

replace OlympusGCP-Firebase => ../OlympusGCP-Firebase

replace OlympusGCP-Intelligence => ../OlympusGCP-Intelligence

replace OlympusGCP-Messaging => ../OlympusGCP-Messaging

replace OlympusGCP-Observability => ../OlympusGCP-Observability

replace OlympusGCP-Storage => ../OlympusGCP-Storage

replace OlympusGCP-Vault => ../OlympusGCP-Vault

replace OlympusGrammar => ../OlympusGrammar

replace OlympusInfrastructure => ../OlympusInfrastructure

replace OlympusVision => ../OlympusVision

replace text => ../Olympus2/00000-Identity-Foundations/P0000-pkg/text

replace pretty => ../Olympus2/00000-Identity-Foundations/P0000-pkg/pretty

replace go-internal => ../Olympus2/00000-Identity-Foundations/P0000-pkg/go-internal

replace check.v1 => ../Olympus2/00000-Identity-Foundations/P0000-pkg/check.v1

replace gopkg.in/check.v1 => ../Olympus2/00000-Identity-Foundations/P0000-pkg/check.v1

require (
	Olympus2 v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)

require gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
