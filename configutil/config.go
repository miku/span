// Package configutil handles application configuration and location and loading of
// various mapping files.
//
// Introduce new config location at /etc/span/assets/..., load data from file,
// not from the binary itself (get rid of go-bindata).
package configutil

// Config is application configuration of span and its subcommands.
type Config struct {
	GitLabToken      string `yaml:"gitlab.token" env:"SPAN_GITLAB_TOKEN"`
	RedmineToken     string `yaml:"redmine.token" env:"SPAN_REDMINE_TOKEN"`
	RedmineURL       string `yaml:"redmine.url" env:"SPAN_REDMINE_URL"`
	WebhookdHostPort string `yaml:"webhookd.listen" env:"SPAN_WEBHOOKD_LISTEN" env-default:"0.0.0.0:8080"`
	WebhookdPath     string `yaml:"webhookd.path" env:"SPAN_WEBHOOKD_PATH" env-default:"trigger"`
	WebhookdLogfile  string `yaml:"webhookd.logfile" env:"SPAN_WEBHOOKD_LOGFILE"`
	WhatIsLiveURL    string `yaml:"whatislive.url" env:"SPAN_WHATISLIVE_URL"`
}
