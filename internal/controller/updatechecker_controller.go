package controller

import (
	"context"
	"strings"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	packagesv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sentinelmetrics "github.com/your-username/olm-update-sentinel/internal/metrics"
)

type SubscriptionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *SubscriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var sub operatorsv1alpha1.Subscription
	if err := r.Get(ctx, req.NamespacedName, &sub); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	pkgName := sub.Spec.Package
	currentChannel := sub.Spec.Channel
	approvalStrategy := string(sub.Spec.InstallPlanApproval)
	installedCSV := sub.Status.InstalledCSV
	currentCSV := sub.Status.CurrentCSV

	// 1. Evaluate update availability on current channel
	updateAvailable := 0.0
	if installedCSV != "" && currentCSV != "" && installedCSV != currentCSV {
		updateAvailable = 1.0
		logger.Info("Update detected", "package", pkgName, "from", installedCSV, "to", currentCSV)
	}

	sentinelmetrics.OperatorUpdateAvailable.WithLabelValues(
		sub.Namespace,
		sub.Name,
		pkgName,
		currentChannel,
		installedCSV,
		currentCSV,
		approvalStrategy,
	).Set(updateAvailable)

	// 2. Fetch PackageManifest to discover newer channels
	var pkgManifest packagesv1.PackageManifest
	pkgErr := r.Get(ctx, types.NamespacedName{Name: pkgName, Namespace: sub.Namespace}, &pkgManifest)
	if pkgErr == nil {
		var channelList []string
		hasNewerChannel := false

		for _, ch := range pkgManifest.Status.Channels {
			channelList = append(channelList, ch.Name)
			if ch.Name != currentChannel {
				hasNewerChannel = true
			}
		}

		newChannelVal := 0.0
		if hasNewerChannel {
			newChannelVal = 1.0
		}

		sentinelmetrics.OperatorNewChannelAvailable.WithLabelValues(
			sub.Namespace,
			sub.Name,
			pkgName,
			currentChannel,
			strings.Join(channelList, ","),
		).Set(newChannelVal)
	}

	return ctrl.Result{}, nil
}

func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorsv1alpha1.Subscription{}).
		Complete(r)
}