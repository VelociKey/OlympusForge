package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"os/signal"

	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	olympusv1 "Olympus2/40000-Communication-Contracts/430-Protocol-Definitions/000-gen/000-olympus/000-v1"
	olympusv1connect "Olympus2/40000-Communication-Contracts/430-Protocol-Definitions/000-gen/000-olympus/000-v1/olympusv1connect"
	mesh "Olympus2/90000-Enablement-Labs/P0000-pkg/000-mesh"
	whisper "Olympus2/90000-Enablement-Labs/P0000-pkg/000-whisper"
)

var Root = getEnv("WORKSPACE_ROOT", "c:\\aAntigravitySpace\\OlympusForge")

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

type ForgeServer struct {
	olympusv1connect.UnimplementedForgeServiceHandler
	busClient    olympusv1connect.EventBusServiceClient
	memoryClient olympusv1connect.MemoryServiceClient
	sc           *whisper.WhisperLog
}

func (s *ForgeServer) Build(ctx context.Context, req *connect.Request[olympusv1.BuildRequest]) (*connect.Response[olympusv1.BuildResponse], error) {
	meta := mesh.FromContext(ctx)
	workspace := req.Msg.Workspace

	slog.Info("⚒️ Forge: Reactive Build Started", "workspace", workspace, "trace_id", meta.TraceID)

	platform := os.Getenv("TARGET_PLATFORM")
	if platform == "" {
		platform = "podman"
	}

	// Dagger Shield: Execute deterministic build pipeline
	forge := &AihubForge{}
	err := forge.Build(ctx, platform, workspace)

	status := "SUCCESS"
	action := "build_success"
	if err != nil {
		status = "FAILURE"
		action = "build_failure"
		slog.Error("⚒️ Forge: Build Failure", "workspace", workspace, "error", err)
	}

	s.memoryClient.LogEvent(ctx, connect.NewRequest(&olympusv1.EventRequest{
		Agent: "Forge", Action: action, Target: workspace, Status: status, Output: fmt.Sprintf("Build target: %s", platform), TraceId: meta.TraceID,
	}))

	return connect.NewResponse(&olympusv1.BuildResponse{
		BuildId: fmt.Sprintf("build-%d", time.Now().Unix()),
		Status:  status,
		Message: fmt.Sprintf("Build target: %s", platform),
	}), nil
}

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))
	meshHubURL := getEnv("MESH_HUB_URL", "http://localhost:8090")
	memoryURL := getEnv("MEMORY_URL", "http://localhost:8084")
	guardianURL := getEnv("GUARDIAN_URL", "http://localhost:8082")
	interceptors := connect.WithInterceptors(mesh.NewInterceptor(guardianURL))
	server := &ForgeServer{
		busClient:    olympusv1connect.NewEventBusServiceClient(http.DefaultClient, memoryURL, interceptors),
		memoryClient: olympusv1connect.NewMemoryServiceClient(http.DefaultClient, memoryURL, interceptors),
		sc:           whisper.New("Forge", "forge.lpsv"),
	}
	mux := http.NewServeMux()
	mux.Handle(olympusv1connect.NewForgeServiceHandler(server, interceptors))
	srv := &http.Server{Addr: ":8088", Handler: h2c.NewHandler(mux, &http2.Server{})}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go subscribeToMutations(server)
	go func() {
		time.Sleep(1 * time.Second)
		mesh.RegisterWithMesh(context.Background(), meshHubURL, "Forge", 8088, "builder", []string{"build", "reactive-ci"})
		srv.ListenAndServe()
	}()
	<-stop
	srv.Shutdown(context.Background())
}

func subscribeToMutations(s *ForgeServer) {
	for {
		stream, err := s.busClient.Subscribe(context.Background(), connect.NewRequest(&olympusv1.SubscribeRequest{AgentName: "Forge", Topic: "mutation"}))
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		for stream.Receive() {
			evt := stream.Msg()
			slog.Info("⚒️ Forge: Mutation detected", "target", evt.Target, "trace_id", evt.TraceId)
			go s.Build(context.Background(), connect.NewRequest(&olympusv1.BuildRequest{Workspace: evt.Target}))
		}
		time.Sleep(2 * time.Second)
	}
}
