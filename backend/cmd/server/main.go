package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hemu1808/auradeploy/backend/internal/api"
	"github.com/hemu1808/auradeploy/backend/internal/cri"
	"github.com/hemu1808/auradeploy/backend/internal/gitops"
	"github.com/hemu1808/auradeploy/backend/internal/network"
	"github.com/hemu1808/auradeploy/backend/internal/observability"
	"github.com/hemu1808/auradeploy/backend/internal/orchestrator"
	"github.com/hemu1808/auradeploy/backend/internal/storage"
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

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	dnsPort := 1053
	if p, err := strconv.Atoi(httpPort); err == nil {
		dnsPort = 1053 + (p - 8080)
	}

	// 2. Initialize Raft Store
	storeNode, err := store.NewStore(nodeID, bindAddr, raftDir, bootstrap)
	if err != nil {
		slog.Error("Failed to initialize Raft store", "error", err)
		os.Exit(1)
	}

	// Wait for leader election (simple sleep for demo, ideally we'd watch leader chan)
	time.Sleep(3 * time.Second)

	// 3. Initialize Network & IPAM (CIDR: 10.244.0.0/16)
	ipam, err := network.NewIPAM(nodeID, "10.244.0.0/16")
	if err != nil {
		slog.Error("Failed to initialize IPAM", "error", err)
	} else {
		// Initialize the bridge and VXLAN on the host (Linux only)
		err = network.SetupNodeNetwork(net.ParseIP("127.0.0.1"), ipam.GetGatewayIP(), ipam.GetNodeSubnet())
		if err != nil {
			slog.Warn("Node CNI setup skipped or failed", "reason", err)
		} else {
			slog.Info("Node CNI setup successful (Bridge/VXLAN overlay)")
		}
	}

	// Start Dummy DNS Server for Service Discovery (Port 1053 to avoid requiring root)
	network.StartDNSServer(dnsPort, logger)

	// 4. Initialize CRI Client (Fallback gracefully if running on Windows without containerd)
	criSock := os.Getenv("CRI_SOCK")
	if criSock == "" {
		criSock = "/run/containerd/containerd.sock"
	}
	criNamespace := os.Getenv("CRI_NAMESPACE")
	if criNamespace == "" {
		criNamespace = "auradeploy"
	}

	var criClient *cri.Client
	criClient, err = cri.NewClient(criSock, criNamespace, ipam, logger)
	if err != nil {
		slog.Warn("Failed to connect to containerd plugin. Falling back to simulation mode.", "error", err, "socket", criSock)
		criClient = nil
	} else {
		slog.Info("Connected to containerd.", "namespace", criNamespace)
	}

	// 5. Initialize Raft Orchestrator
	orch := orchestrator.NewRaftOrchestrator(nodeID, storeNode, criClient, logger)
	slog.Info("Initialized Raft Orchestrator.", "nodeID", nodeID, "addr", bindAddr)

	// 6. Initialize Storage Provisioner (CSI)
	csidefDir := os.Getenv("STORAGE_DIR")
	if csidefDir == "" {
		csidefDir = "./volumes-" + nodeID
	}
	_ = os.MkdirAll(csidefDir, 0755)
	provisioner := storage.NewProvisioner(storeNode, logger, csidefDir)
	provisioner.Start()
	slog.Info("Started Local Volume Provisioner", "dir", csidefDir)

	// 7. Initialize API Handlers
	handler := api.NewHandler(orch)
	gitOpsReconciler := gitops.NewReconciler(orch, logger)
	gitOpsHandler := api.NewGitOpsHandler(gitOpsReconciler)

	// 8. Setup Routing
	mux := http.NewServeMux()

	// Helper to wrap handlers in Auth and RBAC
	authAndRbac := func(resource string, next http.HandlerFunc) http.HandlerFunc {
		return api.RBACMiddleware(orch, resource, next)
	}
	
	// Pre-build the deploy handler wrapped with Admission Webhooks
	deployWithAdmission := api.AdmissionMiddleware(handler.DeployApplicationHandler, api.PodSecurityStandard)

	mux.Handle("/api/v1/applications", api.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Incoming API Request", "method", r.Method, "path", r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			authAndRbac("applications", handler.GetApplicationsHandler)(w, r)
		case http.MethodPost:
			authAndRbac("applications", deployWithAdmission)(w, r)
		case http.MethodPatch:
			authAndRbac("applications", handler.ScaleApplicationHandler)(w, r)
		case http.MethodDelete:
			authAndRbac("applications", handler.RemoveApplicationHandler)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.HandleFunc("/api/v1/cluster/join", handler.JoinClusterHandler)
	mux.HandleFunc("/ws", handler.WebSocketHandler)
	mux.HandleFunc("/api/v1/gitops/sync", gitOpsHandler.WebhookHandler)

	// 9. Attach Observability Metrics
	observability.RegisterClusterMetrics(orch)
	mux.Handle("/metrics", observability.MetricsHandler())

	// 10. Wrap with CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	handlerWithCors := c.Handler(mux)

	// 6. Start Server
	server := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      handlerWithCors,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	slog.Info("AuraDeploy Go Server running", "port", httpPort)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server crashed", "error", err)
	}
}
