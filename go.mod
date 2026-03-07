module olympus.fleet/00SDLC/OlympusForge

go 1.26.0

require (
	connectrpc.com/connect v1.19.1
	dagger.io/dagger v0.20.0
	go.opentelemetry.io/otel v1.40.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/99designs/gqlgen v0.17.81 // indirect
	github.com/Khan/genqlient v0.8.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/vektah/gqlparser/v2 v2.5.32 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/otel/metric v1.40.0 // indirect
	go.opentelemetry.io/otel/sdk v1.40.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.14.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.40.0 // indirect
	go.opentelemetry.io/otel/trace v1.40.0 // indirect
	go.opentelemetry.io/proto/otlp v1.8.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace olympus.fleet/00SDLC/Olympus2/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc => ../Olympus2/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc

replace olympus.fleet/00SDLC/Olympus2/50000-Intelligence-Framework/50200-Logic-Libraries => ../Olympus2/50000-Intelligence-Framework/50200-Logic-Libraries

replace olympus.fleet/00SDLC/Olympus2/70000-Environmental-Harness/dagger => ../Olympus2/70000-Environmental-Harness/70700-Harness-Drivers/dagger-70000

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries

replace olympus.fleet/00SDLC/OlympusFabric/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc => ../OlympusFabric/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc

replace olympus.fleet/00SDLC/OlympusGrammar/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGrammar/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusGCP-Data/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGCP-Data/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusGCP-Firebase/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGCP-Firebase/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusGCP-Storage/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGCP-Storage/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusGCP-Vault/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGCP-Vault/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusGCP-Events/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGCP-Events/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusGCP-Observability/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGCP-Observability/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusGCP-FinOps/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGCP-FinOps/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusGCP-Intelligence/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGCP-Intelligence/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusGCP-Compute/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen => ../OlympusGCP-Compute/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/gen

replace olympus.fleet/00SDLC/OlympusMCP/P0300-Mesh/100-Ground-Substrate => ../OlympusMCP/P0300-Mesh/100-Ground-Substrate

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/110-Auth => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/110-Auth

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/120-Econotel => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/120-Econotel

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/130-ForgeContext => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/130-ForgeContext

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/140-MCPBridge => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/140-MCPBridge

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/150-Mesh => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/150-Mesh

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/170-Policy => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/170-Policy

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/190-Search => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/190-Search

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/200-Substrate => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/200-Substrate

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/210-Vault => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/210-Vault

replace olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/220-Whisper => ../Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/220-Whisper

replace olympus.fleet/00SDLC/OlympusGrammar/parser => ../OlympusGrammar/parser

replace olympus.fleet/10GNDT/GND-Clearinghouse/01000-Identity-Foundations/020-Hardware => ../10GNDT/GND-Clearinghouse/01000-Identity-Foundations/020-Hardware

replace olympus.fleet/10GNDT/GND-Clearinghouse/P0000-pkg/clearing => ../10GNDT/GND-Clearinghouse/P0000-pkg/clearing

replace olympus.fleet/10GNDT/GND-Clearinghouse/P0000-pkg/revenue => ../10GNDT/GND-Clearinghouse/P0000-pkg/revenue

replace olympus.fleet/10GNDT/GND-Customs/01000-Identity-Foundations/020-Hardware => ../10GNDT/GND-Customs/01000-Identity-Foundations/020-Hardware

replace olympus.fleet/10GNDT/GND-Customs/P0000-pkg/compliance => ../10GNDT/GND-Customs/P0000-pkg/compliance

replace olympus.fleet/10GNDT/GND-Freight/P0000-pkg/bale => ../10GNDT/GND-Freight/P0000-pkg/bale

replace olympus.fleet/10GNDT/GND-Freight/P0000-pkg/logistics => ../10GNDT/GND-Freight/P0000-pkg/logistics

replace olympus.fleet/10GNDT/GND-Freight/P0000-pkg/satchel => ../10GNDT/GND-Freight/P0000-pkg/satchel

replace olympus.fleet/10GNDT/GND-Registry/P0000-pkg/indexing => ../10GNDT/GND-Registry/P0000-pkg/indexing

replace olympus.fleet/10GNDT/GND-Substrate/01000-Identity-Foundations/020-Hardware => ../10GNDT/GND-Substrate/01000-Identity-Foundations/020-Hardware

replace olympus.fleet/30INFR/Pinnacle/01000-Identity-Foundations/P0000-pkg/generate => ../30INFR/Pinnacle/01000-Identity-Foundations/P0000-pkg/generate

replace olympus.fleet/30INFR/Pinnacle/01000-Identity-Foundations/P0000-pkg/parse => ../30INFR/Pinnacle/01000-Identity-Foundations/P0000-pkg/parse
