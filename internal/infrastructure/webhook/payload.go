// Package webhook implements the Granola webhook handler.
package webhook

import "time"

// GranolaWebhookPayload is the expected shape of incoming Granola webhook events.
type GranolaWebhookPayload struct {
	Event     string    `json:"event"`
	MeetingID string    `json:"meeting_id"`
	Timestamp time.Time `json:"timestamp"`
}
