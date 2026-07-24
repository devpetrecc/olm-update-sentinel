# olm-update-sentinel
A Kubernetes & OpenShift operator that continuously watches OLM subscriptions, exposes Prometheus metrics for channel updates, and alerts your team on Slack, Teams, and Outlook before versions fall behind.

---

## 🛠️ Quick Start & Setup Guide

To configure notification alerts with `olm-update-sentinel`, you need to:
1. Create a Kubernetes `Secret` containing your credentials and webhook URLs.
2. Apply a `SentinelConfig` Custom Resource (CRD) that references the Secret keys.

---

### Step 1: Create the Secret

Store sensitive webhooks and SMTP credentials safely inside a Kubernetes Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sentinel-notification-secrets
  namespace: olm-update-sentinel-system
type: Opaque
stringData:
  # Slack Webhook URL
  slack-webhook-url: "[https://hooks.slack.com/services/T00/B00/XXXXX](https://hooks.slack.com/services/T00/B00/XXXXX)"
  
  # Microsoft Teams Webhook URL
  teams-webhook-url: "[https://outlook.office.com/webhook/xxx@xxx/IncomingWebhook/xxx](https://outlook.office.com/webhook/xxx@xxx/IncomingWebhook/xxx)"
  
  # Outlook / SMTP Password (or App-Specific Password)
  outlook-smtp-password: "YOUR_APP_SPECIFIC_PASSWORD"
  ```

Apply the secret to your cluster:

```bash
  kubectl apply -f sentinel-secrets.yaml
```

### Step 2: Configure the SentinelConfig CRD

```yaml
apiVersion: sentinel.devpetrecc.io/v1alpha1
kind: SentinelConfig
metadata:
  name: sentinel-config
  namespace: olm-update-sentinel-system
spec:
  # -----------------------------
  # Slack Configuration
  # -----------------------------
  slack:
    enabled: true
    webhookUrlSecret:
      name: sentinel-notification-secrets
      key: slack-webhook-url

  # -----------------------------
  # Microsoft Teams Configuration
  # -----------------------------
  teams:
    enabled: true
    webhookUrlSecret:
      name: sentinel-notification-secrets
      key: teams-webhook-url

  # -----------------------------
  # Outlook / Email Configuration
  # -----------------------------
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

Apply your configuration:

```bash
kubectl apply -f sentinel-config.yaml
```

Note: You can also pass plain-text values directly via webhookUrl or password fields in SentinelConfig during testing or development, though using SecretKeySelector references is strongly recommended for production environments.
