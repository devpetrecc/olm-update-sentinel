[![Go Reference](https://pkg.go.dev/badge/github.com/devpetrecc/olm-update-sentinel.svg)](https://pkg.go.dev/github.com/devpetrecc/olm-update-sentinel)
[![Go Report Card](https://goreportcard.com/badge/github.com/devpetrecc/olm-update-sentinel)](https://goreportcard.com/report/github.com/devpetrecc/olm-update-sentinel)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# olm-update-sentinel

A Kubernetes and OpenShift operator that continuously watches OLM subscriptions, exposes Prometheus metrics for channel updates, and alerts your team via Slack, Teams, and Outlook before versions fall behind.

## Key Features

- **Real-Time Reconciliation:** Continuously monitors OLM `Subscription` and `PackageManifest` objects to detect available version upgrades and channel drifts immediately.
- **Dual Alerting Architecture:** Dispatches direct alerts to chat and email notifications while also exposing native Prometheus gauge metrics for centralized scraping.
- **Production-Grade Security:** Supports dynamic credential resolution via Kubernetes `SecretKeySelector` references, helping avoid plain-text secrets in CRD specs. [web:4]
- **Multi-Channel Dispatcher:** Built-in webhook and SMTP integration for Slack, Microsoft Teams, and Outlook / Office 365.
- **Native Observability:** Seamless integration with Prometheus Operator using standard `ServiceMonitor` and `PrometheusRule` manifests. [web:2][web:7][web:10]

## Architecture Overview

```text
                      ┌──────────────────────────────────────────┐
                      │          OLM Update Sentinel             │
                      │        (Subscription Controller)         │
                      └────────────────────┬─────────────────────┘
                                           │
             ┌─────────────────────────────┼─────────────────────────────┐
             ▼                             ▼                             ▼
┌───────────────────────────┐ ┌──────────────────────--─────┐ ┌───────────────────────────┐
│     Prometheus Metrics    │ │   Kubernetes Secret Resolver│ │  Direct Notification Bus  │
│  (:8080/metrics Endpoint) │ │    (SecretKeySelector)      │ │   (Webhook & SMTP Clients)│
└────────────┬──────────────┘ └────────────┬───────────--───┘ └────────────┬──────────────┘
             │                             │                               │
             ▼                             │                               ▼
┌───────────────────────────┐              │                ┌───────────────────────────┐
│  ServiceMonitor / Rules   │              │                │  Slack / Teams / Outlook  │
│      (Alertmanager)       │              │                │    (Direct Notifications) │
└───────────────────────────┘              │                └───────────────────────────┘
                                           │
                                           ▼
                              ┌───────────────────────────┐
                              │    `sentinel-secrets`     │
                              │     (Opaque Secret)       │
                              └───────────────────────────┘
```

## Quick Start

To configure notification alerts with `olm-update-sentinel`, follow these steps.

### Step 1: Create the Notification Secret

Store sensitive webhooks and SMTP credentials safely inside a Kubernetes Secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sentinel-notification-secrets
  namespace: olm-update-sentinel-system
type: Opaque
stringData:
  slack-webhook-url: "https://hooks.slack.com/services/T00/B00/XXXXX"
  teams-webhook-url: "https://outlook.office.com/webhook/xxx@xxx/IncomingWebhook/xxx"
  outlook-smtp-password: "YOUR_APP_SPECIFIC_PASSWORD"
```

Apply the secret:

```bash
kubectl apply -f sentinel-secrets.yaml
```

### Step 2: Configure the SentinelConfig CRD

Create a `SentinelConfig` custom resource referencing your secret keys.

```yaml
apiVersion: sentinel.devpetrecc.io/v1alpha1
kind: SentinelConfig
metadata:
  name: sentinel-config
  namespace: olm-update-sentinel-system
spec:
  slack:
    enabled: true
    webhookUrlSecret:
      name: sentinel-notification-secrets
      key: slack-webhook-url

  teams:
    enabled: true
    webhookUrlSecret:
      name: sentinel-notification-secrets
      key: teams-webhook-url

  email:
    enabled: true
    smtpHost: "smtp.office365.com"
    smtpPort: 587
    from: "alerts@yourdomain.com"
    to:
      - "devops-team@yourdomain.com"
      - "oncall@yourdomain.com"
    passwordSecret:
      name: sentinel-notification-secrets
      key: outlook-smtp-password
```

Apply the configuration:

```bash
kubectl apply -f sentinel-config.yaml
```

> Note: Direct plain-text fields such as `webhookUrl` and `password` are supported in `SentinelConfig` for development and testing, though using `SecretKeySelector` references is strongly recommended for production environments. [web:4]

## Prometheus Integration

If you prefer routing alerts through Alertmanager instead of direct webhooks, the operator exposes metrics on `:8080/metrics`. `ServiceMonitor` is the standard Prometheus Operator resource for scraping endpoints, and `PrometheusRule` is the standard resource for defining alerting and recording rules. [web:2][web:7][web:10]

### ServiceMonitor Manifest

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: olm-update-sentinel
  namespace: olm-update-sentinel-system
spec:
  selector:
    matchLabels:
      control-plane: controller-manager
  endpoints:
    - port: metrics
      interval: 60s
```

### PrometheusRule Manifest

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: olm-update-sentinel-rules
  namespace: olm-update-sentinel-system
spec:
  groups:
    - name: olm-update-sentinel.rules
      rules:
        - alert: OLMOperatorUpdateAvailable
          expr: olm_operator_update_available == 1
          for: 15m
          labels:
            severity: warning
          annotations:
            summary: "Update available for OLM operator {{ $labels.package }}"
            description: "Package {{ $labels.package }} in namespace {{ $labels.namespace }} can be updated from {{ $labels.installed_csv }} to {{ $labels.current_csv }}."

        - alert: OLMOperatorNewChannelAvailable
          expr: olm_operator_new_channel_available == 1
          for: 1h
          labels:
            severity: info
          annotations:
            summary: "New update channel detected for {{ $labels.package }}"
            description: "Package {{ $labels.package }} in namespace {{ $labels.namespace }} has newer channels available: {{ $labels.channels }}."
```

## Project Structure

Pre-configured, copy-pasteable manifests are available under `config/samples/`:

- `config/samples/sentinel_v1alpha1_sentinelconfig.yaml` — Sample Custom Resource configuration.
- `config/samples/sentinel_secrets.yaml` — Secret template for Slack, Teams, and SMTP passwords.
- `config/samples/servicemonitor.yaml` — Prometheus Operator `ServiceMonitor`.
- `config/samples/prometheusrule.yaml` — Pre-packaged Alertmanager routing rules.
