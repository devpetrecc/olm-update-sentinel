package controller

import (
	"context"
	"strings"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	packagesv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sentinelv1alpha1 "github.com/devpetrecc/olm-update-sentinel/api/v1alpha1"
	sentinelmetrics "github.com/devpetrecc/olm-update-sentinel/internal/metrics"
	"github.com/devpetrecc/olm-update-sentinel/internal/notifier"
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
	isPendingUpdate := installedCSV != "" && currentCSV != "" && installedCSV != currentCSV

	if isPendingUpdate {
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

	// 3. Dispatch Alerts via SentinelConfig CRD if an update is pending
	if isPendingUpdate {
		var configList sentinelv1alpha1.SentinelConfigList
		if err := r.List(ctx, &configList, client.InNamespace(req.Namespace)); err == nil && len(configList.Items) > 0 {
			cfg := configList.Items[0]

			alert := notifier.AlertPayload{
				Title:        "OLM Subscription Update Available",
				Subscription: sub.Name,
				Namespace:    sub.Namespace,
				CurrentCSV:   currentCSV,
				InstalledCSV: installedCSV,
				Channel:      currentChannel,
			}

			// Dispatch Slack
			if cfg.Spec.Slack.Enabled {
				if err := notifier.SendSlack(cfg.Spec.Slack.WebhookURL, alert); err != nil {
					logger.Error(err, "Failed to send Slack alert", "subscription", sub.Name)
				}
			}

			// Dispatch Teams
			if cfg.Spec.Teams.Enabled {
				if err := notifier.SendTeams(cfg.Spec.Teams.WebhookURL, alert); err != nil {
					logger.Error(err, "Failed to send Teams alert", "subscription", sub.Name)
				}
			}

			// Dispatch Email / Outlook
			if cfg.Spec.Email.Enabled {
				var password string
				if cfg.Spec.Email.PasswordSecretRef != "" {
					var sec corev1.Secret
					secKey := types.NamespacedName{Name: cfg.Spec.Email.PasswordSecretRef, Namespace: req.Namespace}
					if err := r.Get(ctx, secKey, &sec); err == nil {
						password = string(sec.Data["password"])
					} else {
						logger.Error(err, "Failed to read Outlook password secret", "secret", cfg.Spec.Email.PasswordSecretRef)
					}
				}

				if err := notifier.SendOutlookSMTP(
					cfg.Spec.Email.SmtpHost,
					cfg.Spec.Email.SmtpPort,
					cfg.Spec.Email.From,
					password,
					cfg.Spec.Email.To,
					alert,
				); err != nil {
					logger.Error(err, "Failed to send Email/Outlook alert", "subscription", sub.Name)
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorsv1alpha1.Subscription{}).
		Complete(r)
}