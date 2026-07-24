package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	OperatorUpdateAvailable = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "olm_sentinel_update_available",
			Help: "Indicates if an update is available on the current channel (1 = update available, 0 = up to date).",
		},
		[]string{"namespace", "subscription", "package", "current_channel", "installed_csv", "latest_csv", "approval_strategy"},
	)

	OperatorNewChannelAvailable = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "olm_sentinel_new_channel_available",
			Help: "Indicates if a newer channel is available in PackageManifest (1 = new channel available, 0 = up to date).",
		},
		[]string{"namespace", "subscription", "package", "current_channel", "available_channels"},
	)
)

func init() {
	metrics.Registry.MustRegister(
		OperatorUpdateAvailable,
		OperatorNewChannelAvailable,
	)
}