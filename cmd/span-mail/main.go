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

	"github.com/miku/span/internal/cmd/mailcmd"
)

var (
	sender   = flag.String("f", "", "The value of the From: header (required)")
	subject  = flag.String("s", "", "The value of the Subject: header (required)")
	textfile = flag.String("b", "", "The textfile for the Body (required)")
	output   = flag.String("o", "", "Print the composed message to FILE instead of sending")
)

// stringSlice implements flag.Value to allow repeated -t flags.
type stringSlice []string

func (s *stringSlice) String() string { return fmt.Sprint(*s) }
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var recipients stringSlice
	flag.Var(&recipients, "t", "A To: header value (at least one required)")
	flag.Parse()
	cfg := mailcmd.Config{
		From:       *sender,
		To:         recipients,
		Subject:    *subject,
		SMTPServer: os.Getenv("SPAN_SMTP_SERVER"),
	}
	// Preserve the original validation order: sender, subject, recipients,
	// then textfile.
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
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
	cfg.Body = string(bodyBytes)
	if *output != "" {
		f, err := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing output file %s: %v\n", *output, err)
			os.Exit(1)
		}
		defer f.Close()
		if err := mailcmd.Run(cfg, f, nil); err != nil {
			fmt.Fprintf(os.Stderr, "error writing output file %s: %v\n", *output, err)
			os.Exit(1)
		}
		return
	}
	if err := mailcmd.Run(cfg, nil, mailcmd.DefaultSender(cfg.SMTPServer)); err != nil {
		fmt.Fprintf(os.Stderr, "error sending mail: %v\n", err)
		os.Exit(1)
	}
}
