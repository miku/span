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
// TODO:
//
// * [ ] proper config handling
// * [ ] send errors into ticket or e-mail
// * [ ] more, maybe more flexible rules
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
	"time"

	"github.com/gorilla/mux"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/miku/span"
	"github.com/miku/span/configutil"
	"github.com/miku/span/gitlab"
	"github.com/miku/span/reviewutil"
	log "github.com/sirupsen/logrus"
)

var (
	addr           = flag.String("addr", "", "hostport to listen on")
	token          = flag.String("token", "", "gitlab auth token, if empty will use span-config")
	repoDir        = flag.String("repo-dir", path.Join(os.TempDir(), "span-webhookd/span"), "local repo clone path")
	logfile        = flag.String("logfile", "", "log to file")
	spanConfigFile = flag.String("span-config", path.Join(span.UserHomeDir(), ".config/span/span.yaml"), "gitlab, redmine tokens, whatislive location")
	triggerPath    = flag.String("trigger-path", "trigger", "path trigger, {host}:{port}/{trigger-path}")
	banner         = fmt.Sprintf(`[<>] webhookd %s`, span.AppVersion)

	// Parsed configuration options.
	config configutil.Config

	// IndexReviewQueue takes requests for index reviews, add some buffering, so we
	// can accept a few requests at a time, although this is improbable.
	IndexReviewQueue = make(chan IndexReviewRequest, 8)
	done             = make(chan bool)
)

// IndexReviewRequest contains information for run an index review.
type IndexReviewRequest struct {
	ReviewConfigFile string
}

// PeekTicketNumber will fish out the ticket number from the YAML review
// configuration. Returns the ticket number (e.g. 1234) and an error.
func (irr *IndexReviewRequest) PeekTicketNumber() (string, error) {
	var (
		reviewConfig = &reviewutil.ReviewConfig{}
		f, err       = os.Open(irr.ReviewConfigFile)
	)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := reviewConfig.ReadFrom(f); err != nil {
		return "", err
	}
	return reviewConfig.Ticket, nil
}

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
		var payload gitlab.PushPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("gitlab payload: %v", payload)
		var (
			pattern     = "^docs/review.*yaml"
			reviewFiles = payload.MatchModified(regexp.MustCompile(pattern))
			repo        = gitlab.Repo{
				URL:   payload.Project.GitHttpUrl,
				Dir:   config.GitLabCloneDir,
				Token: config.GitLabToken,
			}
		)
		if len(reviewFiles) == 0 {
			log.Printf("%s matched nothing, hook done", pattern)
			return
		}
		if err := repo.Update(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// TODO: exit code handling, non-portable, https://stackoverflow.com/a/10385867.
		log.Printf("successfully updated repo at %s", repo.Dir)
		// We can have multiple review files, issue a request for each of them.
		// TODO: the same file might appear multiple times.
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

	var err error
	err = cleanenv.ReadConfig(*spanConfigFile, &config)
	if err != nil {
		err = cleanenv.ReadConfig("/etc/span/span.yaml", &config)
		if err != nil {
			log.Fatalf("failed to read config from: %s and /etc/span/span.yaml", *spanConfigFile)
		}
	}
	// Keep flag compatibility.
	if *token != "" {
		config.GitLabToken = *token
	}
	// Keep flag compatibility.
	if *addr != "" {
		config.WebhookdHostPort = *addr
	}
	// Keep flag compatibility.
	if *repoDir != "" {
		config.GitLabCloneDir = *repoDir
	}

	// Dump config.
	b, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("using config: %s", string(b))

	// Setup handlers.
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc(fmt.Sprintf("/%s", *triggerPath), HookHandler)
	http.Handle("/", r)

	// Log all listening interfaces.
	port, err := parsePort(config.WebhookdHostPort)
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

	// Start background worker.
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

	log.Fatal(http.ListenAndServe(config.WebhookdHostPort, r))
}
