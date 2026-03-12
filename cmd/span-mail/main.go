// span-mail sends an email via SMTP or writes the composed message to a file.
//
// Usage:
//
//	span-mail -f sender@example.com -s "Test subject"
//	          -t recipient1@example.com -t recipient2@example.com
//	          -b body.txt [-o output.txt]
//
// If -o/--output is omitted the message is sent via the SMTP server defined in
// SPAN_SMTP_SERVER (host only; port 25 is used).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/miku/span/mail"
)

// stringSlice implements flag.Value to allow repeated -t flags.
type stringSlice []string

func (s *stringSlice) String() string { return fmt.Sprint(*s) }
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	sender := flag.String("f", "", "The value of the From: header (required)")
	subject := flag.String("s", "", "The value of the Subject: header (required)")
	textfile := flag.String("b", "", "The textfile for the Body (required)")
	output := flag.String("o", "", "Print the composed message to FILE instead of sending")
	var recipients stringSlice
	flag.Var(&recipients, "t", "A To: header value (at least one required)")

	flag.Parse()

	if *sender == "" {
		fmt.Fprintln(os.Stderr, "error: -f/--sender is required")
		flag.Usage()
		os.Exit(1)
	}
	if *subject == "" {
		fmt.Fprintln(os.Stderr, "error: -s/--subject is required")
		flag.Usage()
		os.Exit(1)
	}
	if len(recipients) == 0 {
		fmt.Fprintln(os.Stderr, "error: at least one -t/--recipient is required")
		flag.Usage()
		os.Exit(1)
	}
	if *textfile == "" {
		fmt.Fprintln(os.Stderr, "error: -b/--textfile is required")
		flag.Usage()
		os.Exit(1)
	}

	bodyBytes, err := os.ReadFile(*textfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading body file %s: %v\n", *textfile, err)
		os.Exit(1)
	}

	msg := &mail.Message{
		From:       *sender,
		To:         recipients,
		Subject:    *subject,
		Body:       string(bodyBytes),
		Precedence: "bulk",
	}

	if *output != "" {
		if err := os.WriteFile(*output, msg.Bytes(), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing output file %s: %v\n", *output, err)
			os.Exit(1)
		}
		return
	}

	server := os.Getenv("SPAN_SMTP_SERVER")
	if err := mail.Send(server+":25", nil, msg); err != nil {
		fmt.Fprintf(os.Stderr, "error sending mail: %v\n", err)
		os.Exit(1)
	}
}
