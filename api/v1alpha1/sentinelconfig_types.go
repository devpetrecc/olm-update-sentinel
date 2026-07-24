package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:generate=true
type SlackConfig struct {
	Enabled          bool                      `json:"enabled"`
	WebhookURL       string                    `json:"webhookUrl,omitempty"`
	WebhookURLSecret *corev1.SecretKeySelector `json:"webhookUrlSecret,omitempty"`
}

// +kubebuilder:object:generate=true
type TeamsConfig struct {
	Enabled          bool                      `json:"enabled"`
	WebhookURL       string                    `json:"webhookUrl,omitempty"`
	WebhookURLSecret *corev1.SecretKeySelector `json:"webhookUrlSecret,omitempty"`
}

// +kubebuilder:object:generate=true
type EmailConfig struct {
	Enabled        bool                      `json:"enabled"`
	SmtpHost       string                    `json:"smtpHost,omitempty"`
	SmtpPort       int                       `json:"smtpPort,omitempty"`
	From           string                    `json:"from,omitempty"`
	To             []string                  `json:"to,omitempty"`
	Password       string                    `json:"password,omitempty"`
	PasswordSecret *corev1.SecretKeySelector `json:"passwordSecret,omitempty"`
}

// +kubebuilder:object:generate=true
type SentinelConfigSpec struct {
	Slack SlackConfig `json:"slack,omitempty"`
	Teams TeamsConfig `json:"teams,omitempty"`
	Email EmailConfig `json:"email,omitempty"`
}

// +kubebuilder:object:generate=true
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
	Items           []SentinelConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SentinelConfig{}, &SentinelConfigList{})
}