package notifier

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// Severity levels that map to notification urgency.
const (
	UrgencyLow      = "low"
	UrgencyNormal   = "normal"
	UrgencyCritical = "critical"
)

// Notification represents a desktop notification to send.
type Notification struct {
	Title   string
	Body    string
	Urgency string // "low", "normal", "critical"
	Icon    string // icon name or path (optional)
}

// Notifier sends desktop notifications.
type Notifier struct {
	enabled    bool
	cooldown   time.Duration
	lastNotify time.Time
}

// NewNotifier creates a notifier.
// cooldown prevents notification spam — minimum time between notifications.
func NewNotifier(enabled bool, cooldown time.Duration) *Notifier {
	return &Notifier{
		enabled:  enabled,
		cooldown: cooldown,
	}
}

// Send dispatches a desktop notification.
func (n *Notifier) Send(notif Notification) error {
	if !n.enabled {
		return nil
	}

	// Enforce cooldown
	if time.Since(n.lastNotify) < n.cooldown {
		return nil
	}

	err := sendLinuxNotification(notif)
	if err != nil {
		log.Printf("[notify] failed to send notification: %v", err)
		return err
	}

	n.lastNotify = time.Now()
	log.Printf("[notify] sent: %s", notif.Title)
	return nil
}

// SendDiagnosis is a convenience method for sending diagnosis verdicts as notifications.
func (n *Notifier) SendDiagnosis(severity, title, description string) error {
	urgency := UrgencyNormal
	icon := "network-idle"

	switch severity {
	case "critical":
		urgency = UrgencyCritical
		icon = "network-error"
	case "warning":
		urgency = UrgencyNormal
		icon = "network-offline"
	case "info":
		urgency = UrgencyLow
		icon = "network-idle"
	}

	return n.Send(Notification{
		Title:   fmt.Sprintf("NetPulse: %s", title),
		Body:    description,
		Urgency: urgency,
		Icon:    icon,
	})
}

// sendLinuxNotification uses notify-send on Linux.
func sendLinuxNotification(notif Notification) error {
	args := []string{
		"-u", notif.Urgency,
		"-a", "NetPulse",
	}

	if notif.Icon != "" {
		args = append(args, "-i", notif.Icon)
	}

	// Set expiry: critical stays longer
	switch notif.Urgency {
	case UrgencyCritical:
		args = append(args, "-t", "10000") // 10 seconds
	default:
		args = append(args, "-t", "5000") // 5 seconds
	}

	args = append(args, notif.Title, notif.Body)

	cmd := exec.Command("notify-send", args...)
	return cmd.Run()
}
