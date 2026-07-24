package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SlackConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhookUrl,omitempty"`
}

type TeamsConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhookUrl,omitempty"`
}

type EmailConfig struct {
	Enabled           bool     `json:"enabled"`
	SmtpHost          string   `json:"smtpHost,omitempty"`
	SmtpPort          int      `json:"smtpPort,omitempty"`
	From              string   `json:"from,omitempty"`
	To                []string `json:"to,omitempty"`
	PasswordSecretRef string   `json:"passwordSecretRef,omitempty"`
}

type SentinelConfigSpec struct {
	Slack SlackConfig `json:"slack,omitempty"`
	Teams TeamsConfig `json:"teams,omitempty"`
	Email EmailConfig `json:"email,omitempty"`
}

type SentinelConfigStatus struct {
	LastAlertSent metav1.Time `json:"lastAlertSent,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type SentinelConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SentinelConfigSpec   `json:"spec,omitempty"`
	Status SentinelConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type SentinelConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items             []SentinelConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SentinelConfig{}, &SentinelConfigList{})
}