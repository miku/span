package gitlab

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

// Repo connects a remote git repository to a local clone. Remote
// authentification via token (e.g. in gitlab) supported.
type Repo struct {
	URL   string
	Dir   string
	Token string
}

// AuthURL returns an authenticated repository URL, if no token is supplied,
// just return the repo URL as is.
func (r Repo) AuthURL() string {
	if r.Token == "" {
		return r.URL
	}
	return strings.Replace(r.URL, "https://", fmt.Sprintf("https://oauth2:%s@", r.Token), 1)
}

// String representation.
func (r Repo) String() string {
	return fmt.Sprintf("git repo from %s at %s", r.URL, r.Dir)
}

// Update runs a git pull (or clone), as per strong convention, this will
// always be a fast forward. If repo does not exist yet, clone.
// gitlab/profile/personal_access_tokens: You can also use personal access
// tokens to authenticate against Git over HTTP. They are the only accepted
// password when you have Two-Factor Authentication (2FA) enabled.
func (r Repo) Update() error {
	log.Printf("updating %s", r)
	if r.Token == "" {
		log.Printf("no gitlab.token found, checkout might fail")
	}
	if _, err := os.Stat(path.Dir(r.Dir)); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(r.Dir), 0755); err != nil {
			return err
		}
	}
	var (
		cmd  string
		args []string
	)
	if _, err := os.Stat(r.Dir); os.IsNotExist(err) {
		cmd, args = "git", []string{"clone", r.AuthURL(), r.Dir}
	} else {
		cmd, args = "git", []string{"-C", r.Dir, "pull", "origin", "master"}
	}
	// TODO: black out token for logs.
	log.Printf("[cmd] %s %s", cmd, strings.Join(args, " "))
	// TODO: exit code handling, https://stackoverflow.com/a/10385867.
	return exec.Command(cmd, args...).Run()
}
