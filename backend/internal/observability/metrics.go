package observability

import (
	"net/http"

	"github.com/hemu1808/auradeploy/backend/internal/orchestrator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterClusterMetrics registers cluster-wide metric collectors for Prometheus
func RegisterClusterMetrics(orch orchestrator.Orchestrator) {
	prometheus.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "auradeploy_total_applications",
		Help: "Total number of deployed applications in the cluster",
	}, func() float64 {
		apps, _ := orch.GetApplications()
		return float64(len(apps))
	}))

	prometheus.MustRegister(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auradeploy_deployments_total",
		Help: "Total number of application deployments triggered",
	}))
}

// MetricsHandler returns the standard prometheus HTTP handler
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
