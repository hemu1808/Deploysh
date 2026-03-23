# AuraDeploy: Distributed Orchestrator

AuraDeploy is a high-availability, production-grade container orchestration system designed to abstract away the complexity of Kubernetes while retaining the reliability of Raft consensus, natively integrated OCI-compliant container runtimes, and declarative GitOps pipelines.

This orchestrator is built purely in Go, acting as a sandbox and study of distributed system engineering.

## Overall System Architecture

The overarching architecture of AuraDeploy is split into decoupled sub-components interacting over an internal Finite State Machine (FSM) acting as the single source of truth.

### 1. Raft Consensus Layer (High Availability)
The single-node backend is augmented by an embedded HashiCorp `raft` distributed Key-Value store. The master nodes handle leadership elections, log replications, and robust FSM snapshots. The API handlers gracefully redirect commands to the active leader, providing a true Active-Passive HA Control Plane without relying on an external DB like PostgreSQL or etcd.

### 2. CRI-O / containerd Integration
Rather than executing generic Docker API CLI commands, AuraDeploy acts as a native Container Runtime Interface (CRI) client leveraging `containerd/oci`. It downloads and stages OCI images, mounts network namespaces manually via `netlink`, injects CGroups limits, and utilizes underlying snapshotters for granular layer isolation.

### 3. Custom Scheduler 
A dedicated Go-routine scheduling loop runs on the leader node, monitoring the Raft array for pending Application Replicas lacking a `nodeID`.
*   **Predicates:** Filters nodes based on hard constraints (e.g., `HasSufficientResources`, `VolumeNodeAffinity`).
*   **Priorities:** Scores mathematically viable nodes (e.g., `LeastAllocated` algorithm to ensure an optimal spread of CPU/Memory footprints across the cluster).
Selected Node targets are broadcast into the Raft log for respective worker nodes to claim and execute.

### 4. Custom Network Overlay (CNI)
Operating natively on Linux environments, worker nodes establish a robust multi-host networking mesh interface.
*   **IPAM:** Deterministic `/24` subnet slicing out of an ephemeral cluster-wide `/16` CIDR.
*   **Linux Bridge & VXLAN:** Inter-node communication operates by attaching physical Host Bridges (`aura0`) to VXLAN overlays (`vxlan0`).
*   **Network Namespaces:** The CRI worker isolates active containers via isolated `veth` pairs spliced between the host bridge and the container namespace.
*   **Service Discovery:** Applications register against an integrated dummy UDP DNS server reacting dynamically to FSM placement updates to yield service resolution.

### 5. Custom CSI (Storage Provisioner)
AuraDeploy executes its own Container Storage Interface logic to persist file architectures outside the volatile lifecycle of transient containers.
*   **Local Volume Controller:** An embedded reconciliation loop that provisions localized host-paths mapping 1:1 against Persistent Volume Claims (PVCs) requesting nodes.
*   **Volume Binding:** Containers get fully resolved OCI Bind Mounts injected smoothly into their spec templates by the CRI daemon.

### 6. API Security & Admission Control
A robust, internally native role-managed security firewall restricts the API multiplexer logic.
*   **Authentication Middleware:** Generic JWT Bearer token extraction and user identification.
*   **RBAC Mapping:** Authorizes verified Subjects against defined FSM `Roles` & `RoleBindings` determining specific access rights natively (comparing HTTP Verbs against allowed Resource Endpoints).
*   **Admission Controller:** Validating Webhooks intercept active deployments to strip Privileged Containers, immediately rejecting payloads that attempt to attach to the root filesystem or override the `RUN_AS_ROOT` PodSecurityStandard constraints.

### 7. GitOps Declarative Reconciler
Applications can be managed imperatively using localized UI/API operations, or pulled directly declaratively via the `GitOps` Engine.
*   **YAML Parser:** Intelligently reads Kubernetes-style custom definitions mapping `kind: Application` schemas using `yaml.v3`.
*   **Drift Check:** Actively diff-checks recorded local cluster states versus remote declarative definitions, enforcing immediate redeployments to heal drifted ENV vars, corrupt images, or invalid replica settings seamlessly.

### 8. Observability Stack
Visibility traces are embedded directly into the core networking loops.
*   **Prometheus Metrics:** A generalized prometheus HTTP exporter exposing custom FSM Gauges (e.g., `auradeploy_total_applications` / deployments counted).
*   **OpenTelemetry:** Distributed structural context traces registered into the routing pipeline structure (`go.opentelemetry.io/otel`).
*   **Event Recorder:** Pushing structured historical cluster-mutations over standard I/O (simulating Kubernetes specific `Events`).

---

## Requirements

*   **Go** (1.20+)
*   **Node.js** (v18+) & **npm**

## Getting Started

### 1. Start the Backend (AuraDeploy FSM Leader)

Open a terminal and navigate to the `backend` directory to run the seed node:

```bash
cd backend
$env:RAFT_NODE_ID="node1"
$env:RAFT_BIND_ADDR="127.0.0.1:9000"
$env:PORT="8080"
$env:RAFT_BOOTSTRAP="true"
go run ./cmd/server
```

**Join a Worker Node:** (Optional)
Open another terminal:
```bash
cd backend
$env:RAFT_NODE_ID="node2"
$env:RAFT_BIND_ADDR="127.0.0.1:9001"
$env:PORT="8081"
go run ./cmd/server -join 127.0.0.1:8080
```

### 2. Start the Frontend (Dashboard)

Open a new terminal at the root of the project to run the React dashboard:

```bash
npm install
npm run dev
```

## Verification & Testing

To verify the platform is running successfully:

1.  **Dashboard UI:** Open your browser to [http://localhost:5173](http://localhost:5173). You should see the AuraDeploy dashboard.
2.  **API Health / Cluster State:** Use `curl` to check the backend leader's state directly:
    ```bash
    # (Replace endpoint with actual status endpoint if known, e.g. /status or check applications)
    curl http://localhost:8080/applications
    ```

###
I built AuraDeploy to challenge the assumption that container orchestration requires heavyweight control planes. While Kubernetes dominates production, its complexity creates massive operational overhead for smaller deployments. I wanted to prove that a single Go binary with SQLite could achieve the same core guarantees—self-healing, real-time state synchronization, and declarative scaling—through careful systems design rather than distributed consensus. The project explores how far you can push a monolithic architecture before hitting fundamental limits.
