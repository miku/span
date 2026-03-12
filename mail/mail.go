// Package mail provides simple email composition and sending via SMTP.
package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
)

// Message represents an email message.
type Message struct {
	From       string
	To         []string
	Subject    string
	Body       string
	Precedence string // e.g. "bulk"; left empty if not needed
}

// Bytes composes the raw RFC 2822 message.
func (m *Message) Bytes() []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", m.From)
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(m.To, ", "))
	fmt.Fprintf(&buf, "Subject: %s\r\n", m.Subject)
	if m.Precedence != "" {
		fmt.Fprintf(&buf, "Precedence: %s\r\n", m.Precedence)
	}
	buf.WriteString("\r\n")
	buf.WriteString(m.Body)
	return buf.Bytes()
}

// Send sends the message via the given SMTP server address (host:port).
// Auth may be nil for unauthenticated relays.
func Send(addr string, auth smtp.Auth, msg *Message) error {
	return smtp.SendMail(addr, auth, msg.From, msg.To, msg.Bytes())
}
