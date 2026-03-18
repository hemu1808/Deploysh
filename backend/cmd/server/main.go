package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/hemu1808/auradeploy/backend/internal/api"
	"github.com/hemu1808/auradeploy/backend/internal/cri"
	"github.com/hemu1808/auradeploy/backend/internal/orchestrator"
	"github.com/hemu1808/auradeploy/backend/internal/store"
	"github.com/rs/cors"
)

func main() {
	// 0. Setup Structured Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	
	slog.Info("Starting AuraDeploy Backend initialization...")

	// 1. Parse configuration flags (Basic setup, to be moved to cobra later ideally)
	nodeID := os.Getenv("RAFT_NODE_ID")
	if nodeID == "" {
		nodeID = "node1"
	}
	bindAddr := os.Getenv("RAFT_BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "127.0.0.1:5000"
	}
	raftDir := os.Getenv("RAFT_DIR")
	if raftDir == "" {
		raftDir = "./raft-data-" + nodeID
	}
	bootstrapEnv := os.Getenv("RAFT_BOOTSTRAP")
	bootstrap := true
	if bootstrapEnv != "" {
		bootstrap = bootstrapEnv == "true"
	}

	// 2. Initialize Raft Store
	storeNode, err := store.NewStore(nodeID, bindAddr, raftDir, bootstrap)
	if err != nil {
		slog.Error("Failed to initialize Raft store", "error", err)
		os.Exit(1)
	}

	// Wait for leader election (simple sleep for demo, ideally we'd watch leader chan)
	time.Sleep(3 * time.Second)

	// 3. Initialize CRI Client (Fallback gracefully if running on Windows without containerd)
	criSock := os.Getenv("CRI_SOCK")
	if criSock == "" {
		criSock = "/run/containerd/containerd.sock"
	}
	criNamespace := os.Getenv("CRI_NAMESPACE")
	if criNamespace == "" {
		criNamespace = "auradeploy"
	}

	var criClient *cri.Client
	criClient, err = cri.NewClient(criSock, criNamespace, logger)
	if err != nil {
		slog.Warn("Failed to connect to containerd plugin. Falling back to simulation mode.", "error", err, "socket", criSock)
		criClient = nil
	} else {
		slog.Info("Connected to containerd.", "namespace", criNamespace)
	}

	// 4. Initialize Raft Orchestrator
	orch := orchestrator.NewRaftOrchestrator(storeNode, criClient, logger)
	slog.Info("Initialized Raft Orchestrator.", "nodeID", nodeID, "addr", bindAddr)

	// 4. Initialize API Handlers
	handler := api.NewHandler(orch)

	// 5. Setup Routing
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/applications", func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Incoming API Request", "method", r.Method, "path", r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			handler.GetApplicationsHandler(w, r)
		case http.MethodPost:
			handler.DeployApplicationHandler(w, r)
		case http.MethodPatch:
			handler.ScaleApplicationHandler(w, r)
		case http.MethodDelete:
			handler.RemoveApplicationHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/cluster/join", handler.JoinClusterHandler)
	mux.HandleFunc("/ws", handler.WebSocketHandler)

	// 5. Wrap with CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	handlerWithCors := c.Handler(mux)

	// 6. Start Server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handlerWithCors,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	slog.Info("AuraDeploy Go Server running", "port", 8080)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server crashed", "error", err)
	}
}
