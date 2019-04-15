// span-webhookd can serve as a webhook receiver[1] for gitlab, refs #13499.
//
// We listen for push hooks to trigger index reviews via span-review.
//
// [1] https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#example-webhook-receiver
//
// Configuration (Redmine, Gitlab, Index), by default in
// ~/.config/span/span.json, fallback at /etc/span/span.json. This config file
// is used both by span-webhookd and span-review.
//
// {
//    "gitlab.token": "g0d8gf0LKJWg89dsf8gd0gf9-YU",
//    "whatislive.url": "http://example.com/whatislive",
//    "redmine.baseurl": "https://projects.example.com",
//    "redmine.apitoken": "badfb87ab7987daafbd9db",
//    "port": 8080
// }
//
// Some limitations:
//
// * By default, the server will listen on all interfaces, only the port number
//   is configurable.
// * There is no error reporting except in the logs.
// * Exit code from spawned span-review is ignored.
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/miku/span"
	"github.com/miku/span/reviewutil"
	log "github.com/sirupsen/logrus"
)

// DefaultPort to listen on for gitlab hook.
const DefaultPort = 8080

var (
	addr           = flag.String("addr", fmt.Sprintf(":%d", findConfiguredPort()), "hostport to listen on")
	token          = flag.String("token", "", "gitlab auth token, if empty will use span-config")
	repoDir        = flag.String("repo-dir", path.Join(os.TempDir(), "span-webhookd/span"), "local repo clone path")
	logfile        = flag.String("logfile", "", "log to file")
	spanConfigFile = flag.String("span-config", path.Join(span.UserHomeDir(), ".config/span/span.json"), "gitlab, redmine tokens, whatislive location")
	triggerPath    = flag.String("trigger-path", "trigger", "path trigger, {host}:{port}/{trigger-path}")
	banner         = fmt.Sprintf(`[<>] webhookd %s`, span.AppVersion)
)

// IndexReviewRequest contains information for run an index review.
type IndexReviewRequest struct {
	ReviewConfigFile string
}

// PeekTicketNumber will fish out the ticket number from the YAML review
// configuration.
func (irr *IndexReviewRequest) PeekTicketNumber() (ticket string, err error) {
	reviewConfig := &reviewutil.ReviewConfig{}
	f, err := os.Open(irr.ReviewConfigFile)
	if err != nil {
		return ticket, err
	}
	defer f.Close()
	if _, err := reviewConfig.ReadFrom(f); err != nil {
		return ticket, err
	}
	return reviewConfig.Ticket, nil
}

// IndexReviewQueue takes requests for index reviews, add come buffering, so we
// can accept a few requests at a time, although this is improbable.
var IndexReviewQueue = make(chan IndexReviewRequest, 100)
var done = make(chan bool)

// Worker hangs in there, checks for any new review requests on the index
// review queue and starts review, if requested.
func Worker(done chan bool) {
	log.Println("worker started")
	for rr := range IndexReviewQueue {
		log.Printf("worker received review request: %s, running review ...", rr)

		cmd, args := "span-review", []string{"-c", rr.ReviewConfigFile}
		log.Printf("[cmd] %s %s", cmd, strings.Join(args, " "))

		// TODO(miku): use runCommand, but handle stdout, stderr as well.
		out, err := exec.Command(cmd, args...).CombinedOutput()
		if err != nil {
			log.Printf("%s failed: %s, combined output: %s", cmd, err, string(out))
			continue
		}
		// TODO(miku): use PeekTicketNumber to send error to ticket if needed.
		log.Printf("[output] %s", out)
		log.Println("review completed")
	}
	log.Println("worker shutdown")
	done <- true
}

// runCommand will execute a program with args and will return an error, if the
// program returned an non-zero exit code. The error can then be propagated
// into logs or a ticket. This works on Linux only (and maybe, accidentally on
// Windows), https://stackoverflow.com/a/10385867.
func runCommand(program string, args ...string) error {
	cmd := exec.Command(program, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("[runCommand] could not start: %v", cmd)
	}
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("[runCommand] %v exit status: %d", cmd.Path, status.ExitStatus())
			}
		} else {
			fmt.Errorf("[runCommand] cmd.Wait failed for %v", cmd)
		}
	}
	return nil
}

// Repo points to a local clone containing the review configuration we want. A
// personal access token is required to clone the repo from GitLab.
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
		log.Printf("warning: not gitlab.token found, checkout might fail (%s)", *spanConfigFile)
	}
	if _, err := os.Stat(path.Dir(r.Dir)); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(r.Dir), 0755); err != nil {
			return err
		}
	}

	var cmd string
	var args []string

	if _, err := os.Stat(r.Dir); os.IsNotExist(err) {
		cmd, args = "git", []string{"clone", r.AuthURL(), r.Dir}
	} else {
		cmd, args = "git", []string{"-C", r.Dir, "pull", "origin", "master"}
	}
	// XXX: black out token for logs.
	log.Printf("[cmd] %s %s", cmd, strings.Join(args, " "))
	// XXX: exit code handling, https://stackoverflow.com/a/10385867.
	return exec.Command(cmd, args...).Run()
}

// PushPayload delivered on push and web edits. This is the whole response, we
// are mainly interested in the modified files in a commit.
type PushPayload struct {
	After       string `json:"after"`
	Before      string `json:"before"`
	CheckoutSha string `json:"checkout_sha"`
	Commits     []struct {
		Added  []interface{} `json:"added"`
		Author struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"author"`
		Id        string        `json:"id"`
		Message   string        `json:"message"`
		Modified  []string      `json:"modified"`
		Removed   []interface{} `json:"removed"`
		Timestamp string        `json:"timestamp"`
		Url       string        `json:"url"`
	} `json:"commits"`
	EventName  string      `json:"event_name"`
	Message    interface{} `json:"message"`
	ObjectKind string      `json:"object_kind"`
	Project    struct {
		AvatarUrl         interface{} `json:"avatar_url"`
		CiConfigPath      interface{} `json:"ci_config_path"`
		DefaultBranch     string      `json:"default_branch"`
		Description       string      `json:"description"`
		GitHttpUrl        string      `json:"git_http_url"`
		GitSshUrl         string      `json:"git_ssh_url"`
		Homepage          string      `json:"homepage"`
		HttpUrl           string      `json:"http_url"`
		Id                int64       `json:"id"`
		Name              string      `json:"name"`
		Namespace         string      `json:"namespace"`
		PathWithNamespace string      `json:"path_with_namespace"`
		SshUrl            string      `json:"ssh_url"`
		Url               string      `json:"url"`
		VisibilityLevel   int64       `json:"visibility_level"`
		WebUrl            string      `json:"web_url"`
	} `json:"project"`
	ProjectId  int64  `json:"project_id"`
	Ref        string `json:"ref"`
	Repository struct {
		Description     string `json:"description"`
		GitHttpUrl      string `json:"git_http_url"`
		GitSshUrl       string `json:"git_ssh_url"`
		Homepage        string `json:"homepage"`
		Name            string `json:"name"`
		Url             string `json:"url"`
		VisibilityLevel int64  `json:"visibility_level"`
	} `json:"repository"`
	TotalCommitsCount int64  `json:"total_commits_count"`
	UserAvatar        string `json:"user_avatar"`
	UserEmail         string `json:"user_email"`
	UserId            int64  `json:"user_id"`
	UserName          string `json:"user_name"`
	UserUsername      string `json:"user_username"`
}

// ModifiedFiles returns all modified files across all commits in this payload.
func (p PushPayload) ModifiedFiles() (filenames []string) {
	for _, commit := range p.Commits {
		for _, modified := range commit.Modified {
			filenames = append(filenames, modified)
		}
	}
	return
}

// IsFileModified returns true, if given file has been modified.
func (p PushPayload) IsFileModified(filename string) bool {
	for _, modified := range p.ModifiedFiles() {
		if modified == filename {
			return true
		}
	}
	return false
}

// MatchModified returns a list of paths matching a pattern (match against the
// full path in repo, e.g. docs/review.*).
func (p PushPayload) MatchModified(re *regexp.Regexp) (filenames []string) {
	for _, modified := range p.ModifiedFiles() {
		if re.MatchString(modified) {
			filenames = append(filenames, modified)
			log.Printf("%s matches %s", modified, re)
		} else {
			log.Printf("%s ignored", modified)
		}
	}
	return
}

// HookHandler can act as webhook receiver. The hook we use at the moment is
// the Push Hook. Other types are Issue, Note or Tag Push Hook.
func HookHandler(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	defer func() {
		// We care a bit, because gitlab wants us to return ASAP.
		log.Printf("request completed after %s", time.Since(started))
	}()

	log.Printf("request from %s", r.RemoteAddr)
	if r.Header.Get("X-FORWARDED-FOR") != "" {
		log.Printf("X-FORWARDED-FOR: %s", r.Header.Get("X-FORWARDED-FOR"))
	}

	gitlabEvent := strings.TrimSpace(r.Header.Get("X-Gitlab-Event"))

	switch gitlabEvent {
	case "Push Hook":
		var payload PushPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("gitlab payload: %v", payload)

		pattern := "^docs/review.*yaml"
		reviewFiles := payload.MatchModified(regexp.MustCompile(pattern))
		if len(reviewFiles) == 0 {
			log.Printf("%s matched nothing, hook done", pattern)
			return
		}
		repo := Repo{
			URL:   payload.Project.GitHttpUrl,
			Dir:   *repoDir,
			Token: *token,
		}
		if err := repo.Update(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// XXX: exit code handling, non-portable, https://stackoverflow.com/a/10385867.
		log.Printf("successfully updated repo at %s", repo.Dir)

		// We can have multiple review files, issue a request for each of them.
		for _, reviewFile := range reviewFiles {
			rr := IndexReviewRequest{
				ReviewConfigFile: path.Join(repo.Dir, reviewFile),
			}
			IndexReviewQueue <- rr
			log.Printf("dispatched review for %s", reviewFile)
		}
	case "":
		log.Printf("X-Gitlab-Event not given or empty")
		w.WriteHeader(http.StatusBadRequest)
	default:
		log.Printf("unregistered or invalid event kind: %s", gitlabEvent)
		w.WriteHeader(http.StatusBadRequest)
	}
}

// HomeHandler says hello.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	s := fmt.Sprintf("This is span-webhookd %s, a webhook receiver for gitlab (#12756).", span.AppVersion)
	if _, err := fmt.Fprintln(w, s); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// parsePort takes a hostport and returns the port number as int.
func parsePort(addr string) (int, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("cannot parse port")
	}
	return strconv.Atoi(parts[1])
}

// findGitlabToken returns the GitLab auth token, if configured. XXX: Use
// proper config library.
func findGitlabToken() (string, error) {
	if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
		// XXX: Use a real config framework, not these hacks.
		*spanConfigFile = "/etc/span/span.json"
		if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
			return "", err
		}
	}
	f, err := os.Open(*spanConfigFile)
	if err != nil {
		return "", err
	}
	var conf struct {
		Token string `json:"gitlab.token"`
	}
	if err := json.NewDecoder(f).Decode(&conf); err != nil {
		return "", err
	}
	return conf.Token, nil
}

// findConfiguredPort returns a configured port number or 8080 if none is
// specified. XXX: Use proper config library.
func findConfiguredPort() int {
	if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
		return DefaultPort
	}
	f, err := os.Open(*spanConfigFile)
	if err != nil {
		return DefaultPort
	}
	var conf struct {
		Port int `json:"port"`
	}
	if err := json.NewDecoder(f).Decode(&conf); err != nil {
		return DefaultPort
	}
	if conf.Port == 0 {
		conf.Port = DefaultPort
	}
	return conf.Port
}

func main() {
	flag.Parse()

	log.Println(banner)

	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Printf("starting GitLab webhook receiver (%s) on %s ... (settings/integrations)",
		span.AppVersion, *addr)

	// Fallback configuration, since daemon home is /usr/sbin.
	if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
		*spanConfigFile = "/etc/span/span.json"
	}
	// XXX: Use a real framework like go-ucfg or globalconf.
	if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
		log.Fatal("no configuration found, put one into /etc/span/span.json")
	}
	log.Printf("using fallback configuration from %s", *spanConfigFile)

	var err error

	if *token == "" {
		*token, err = findGitlabToken()
		if err != nil {
			log.Printf("gitlab token not configured, might be an issue")
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc(fmt.Sprintf("/%s", *triggerPath), HookHandler)
	http.Handle("/", r)

	port, err := parsePort(*addr)
	if err != nil {
		log.Fatal(err)
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok {
			log.Printf("http://%s:%d/%s", ipnet.IP.String(), port, *triggerPath)
		}
	}

	go Worker(done)
	log.Println("use CTRL-C to gracefully stop server")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		// XXX: Use some timeout here.
		for range c {
			close(IndexReviewQueue)
			<-done
			os.Exit(0)
		}
	}()

	log.Fatal(http.ListenAndServe(*addr, r))
}
