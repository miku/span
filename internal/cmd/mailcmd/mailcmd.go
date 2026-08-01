// Package mailcmd composes an email message and either writes it to a writer or
// sends it via an injected sender. It backs the span-mail command.
package mailcmd

import (
	"fmt"
	"io"

	"github.com/miku/span/mail"
)

// Config holds the fields needed to compose a message. Body is the already-read
// message body (main reads the -b file); SMTPServer is the host part of the
// server address ("" if a composed-message writer is used instead).
type Config struct {
	From       string
	To         []string
	Subject    string
	Body       string
	SMTPServer string
}

// Sender sends a composed message. It is injected so Run can be tested without
// touching the network; the default sender talks SMTP.
type Sender func(msg *mail.Message) error

// DefaultSender sends the message via SMTP to server:25 with no auth.
func DefaultSender(server string) Sender {
	return func(msg *mail.Message) error {
		return mail.Send(server+":25", nil, msg)
	}
}

// Validate reports whether the required fields are present.
func (c Config) Validate() error {
	switch {
	case c.From == "":
		return fmt.Errorf("-f/--sender is required")
	case c.Subject == "":
		return fmt.Errorf("-s/--subject is required")
	case len(c.To) == 0:
		return fmt.Errorf("at least one -t/--recipient is required")
	}
	return nil
}

// message builds a mail.Message from the config.
func (c Config) message() *mail.Message {
	return &mail.Message{
		From:       c.From,
		To:         c.To,
		Subject:    c.Subject,
		Body:       c.Body,
		Precedence: "bulk",
	}
}

// Run composes the message. If w is non-nil the composed message is written to
// it (the -o path); otherwise it is handed to send.
func Run(cfg Config, w io.Writer, send Sender) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	msg := cfg.message()
	if w != nil {
		_, err := w.Write(msg.Bytes())
		return err
	}
	if send == nil {
		return fmt.Errorf("no output writer and no sender configured")
	}
	return send(msg)
}
