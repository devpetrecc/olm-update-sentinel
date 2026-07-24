package notifier

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"time"
)

type AlertPayload struct {
	Title        string
	Subscription string
	Namespace    string
	CurrentCSV   string
	InstalledCSV string
	Channel      string
}

// --- SLACK ---
type SlackPayload struct {
	Text string `json:"text"`
}

func SendSlack(webhookURL string, alert AlertPayload) error {
	msg := fmt.Sprintf("🚨 *%s*\n*Subscription:* `%s`\n*Namespace:* `%s`\n*Channel:* `%s`\n*New CSV:* `%s`",
		alert.Title, alert.Subscription, alert.Namespace, alert.Channel, alert.CurrentCSV)
	return postJSON(webhookURL, SlackPayload{Text: msg})
}

// --- TEAMS ---
type TeamsMessage struct {
	Type       string `json:"@type"`
	Context    string `json:"@context"`
	Summary    string `json:"summary"`
	Text       string `json:"text"`
}

func SendTeams(webhookURL string, alert AlertPayload) error {
	msg := fmt.Sprintf("⚠️ **%s**<br/>**Subscription:** %s<br/>**Namespace:** %s<br/>**New CSV:** %s",
		alert.Title, alert.Subscription, alert.Namespace, alert.CurrentCSV)
	return postJSON(webhookURL, TeamsMessage{
		Type:    "MessageCard",
		Context: "https://schema.org/extensions",
		Summary: alert.Title,
		Text:    msg,
	})
}

// --- OUTLOOK / SMTP ---
func SendOutlookSMTP(host string, port int, from string, password string, to []string, alert AlertPayload) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	subject := fmt.Sprintf("Subject: [OLM Alert] Update Available for %s\n", alert.Subscription)
	body := fmt.Sprintf("Content-Type: text/html; charset=UTF-8\n\n<h2>⚠️ OLM Update Available</h2><p><b>Subscription:</b> %s</p><p><b>Namespace:</b> %s</p><p><b>Target CSV:</b> %s</p>",
		alert.Subscription, alert.Namespace, alert.CurrentCSV)

	auth := smtp.PlainAuth("", from, password, host)
	tlsconfig := &tls.Config{ServerName: host}

	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.StartTLS(tlsconfig); err != nil {
		return err
	}
	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	for _, rec := range to {
		if err = client.Rcpt(rec); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(subject + body))
	if err != nil {
		return err
	}
	return w.Close()
}

func postJSON(url string, payload interface{}) error {
	jsonVal, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonVal))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http error %d", resp.StatusCode)
	}
	return nil
}