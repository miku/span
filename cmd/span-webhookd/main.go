// span-webhookd can serve as a webhook receiver[1] for gitlab, refs #13499.
//
// We listen for push hooks to trigger index reviews.
//
// [1] https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#example-webhook-receiver
//
// Configuration (Redmine, Gitlab, Index) is expected in ~/.config/span/span.json.
//
// {
//    "gitlab.token": "g0d8gf0LKJWg89dsf8gd0gf9-YU",
//    "whatislive.url": "http://example.com/whatislive",
//    "redmine.baseurl": "https://projects.example.com",
//    "redmine.apitoken": "badfb87ab7987daafbd9db"
// }

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/miku/span"
	log "github.com/sirupsen/logrus"
)

const DefaultPort = 8080

var (
	addr           = flag.String("addr", fmt.Sprintf(":%d", findConfiguredPort()), "hostport to listen on")
	token          = flag.String("token", "", "gitlab auth token, if empty will use span-config")
	repoDir        = flag.String("repo-dir", path.Join(os.TempDir(), "span-webhookd/span"), "local repo clone path")
	logfile        = flag.String("logfile", "", "log to file")
	spanConfigFile = flag.String("span-config", path.Join(UserHomeDir(), ".config/span/span.json"), "gitlab, redmine tokens, whatislive location")
	banner         = `
                         888       888                        888   _         888
Y88b    e    /  e88~~8e  888-~88e  888-~88e  e88~-_   e88~-_  888 e~ ~   e88~\888
 Y88b  d8b  /  d888  88b 888  888b 888  888 d888   i d888   i 888d8b    d888  888
  Y888/Y88b/   8888__888 888  8888 888  888 8888   | 8888   | 888Y88b   8888  888
   Y8/  Y8/    Y888    , 888  888P 888  888 Y888   ' Y888   ' 888 Y88b  Y888  888
    Y    Y      "88___/  888-_88"  888  888  "88_-~   "88_-~  888  Y88b  "88_/888
`
)

// IndexReviewRequest contains information for run an index review.
type IndexReviewRequest struct {
	ReviewConfigFile string
}

// IndexReviewQueue takes requests for index reviews.
var IndexReviewQueue = make(chan IndexReviewRequest, 100)
var done = make(chan bool)

// Worker hangs in there, checks for any new review requests and starts to run
// the review, if required..
func Worker(done chan bool) {
	log.Println("worker started")
	for rr := range IndexReviewQueue {
		log.Printf("worker received review request: %s, running review ...", rr)

		cmd, args := "span-review", []string{"-c", rr.ReviewConfigFile}
		log.Printf("[cmd] %s %s", cmd, strings.Join(args, " "))

		out, err := exec.Command(cmd, args...).CombinedOutput() // XXX: Pick off exit code.
		if err != nil {
			log.Printf("%s failed: %s, combined output: %s", cmd, err, string(out))
			continue
		}
		log.Printf("[output] %s", out)
		log.Println("review completed")
	}
	log.Println("worker shutdown")
	done <- true
}

// Repo points to a local clone containing the configuration we want.
type Repo struct {
	URL   string
	Dir   string
	Token string
}

// AuthURL returns an authenticated repository URL.
func (r Repo) AuthURL() string {
	return strings.Replace(r.URL, "https://", fmt.Sprintf("https://oauth2:%s@", r.Token), 1)
}

// String representation.
func (r Repo) String() string {
	return fmt.Sprintf("Repo from %s at %s", r.URL, r.Dir)
}

// Update just runs a git pull, as per strong convention, this will always be a
// fast forward. If repo does not exist yet, clone.
func (r Repo) Update() error {
	log.Printf("updating %s", r)
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
	log.Printf("[cmd] %s %s", cmd, strings.Replace(strings.Join(args, " "), r.Token, "xxxxxxxx", -1))
	return exec.Command(cmd, args...).Run()
}

// PushPayload delivered on push and web edits.
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

// HookHandler can act as webhook receiver. The hook we use at the moment is
// the Push Hook. Other types are Issue, Note or Tag Push Hook.
func HookHandler(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	defer func() {
		// We care a bit, because gitlab wants us to return ASAP.
		log.Printf("request completed after %s", time.Since(started))
	}()

	gitlabEvent := strings.TrimSpace(r.Header.Get("X-Gitlab-Event"))
	switch gitlabEvent {
	case "Push Hook":
		var payload PushPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("gitlab payload: %v", payload)
		log.Printf("modified files: %s", strings.Join(payload.ModifiedFiles(), ", "))
		if !payload.IsFileModified("docs/review.yaml") {
			log.Println("review.yaml not modified, hook done")
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

		// XXX: exit code handling, non-portable.
		log.Printf("successfully updated repo at %s", repo.Dir)

		rr := IndexReviewRequest{
			ReviewConfigFile: path.Join(repo.Dir, "docs/review.yaml"),
		}
		IndexReviewQueue <- rr
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

// UserHomeDir returns the home directory of the user. XXX: Factor this out.
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func parsePort(addr string) (int, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("cannot parse port")
	}
	return strconv.Atoi(parts[1])
}

// findGitlabToken returns the GitLab auth token, if configured.
func findGitlabToken() (string, error) {
	if _, err := os.Stat(*spanConfigFile); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(*spanConfigFile), 0755); err != nil {
			return "", err
		}
		data := []byte(`{"gitlab.token": "xxx"}`)
		if err := ioutil.WriteFile(*spanConfigFile, data, 0600); err != nil {
			return "", err
		}
		return "", fmt.Errorf("created new config file, please adjust: %s", *spanConfigFile)
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

// findConfiguredPort returns a configured port number or 8080 if none is specified.
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

	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	var err error

	if *token == "" {
		*token, err = findGitlabToken()
	}

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/trigger", HookHandler)
	http.Handle("/", r)

	log.Println(banner)
	log.Printf("starting GitLab webhook receiver (%s) on %s ... (settings/integrations)",
		span.AppVersion, *addr)

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
			log.Printf("http://%s:%d/trigger", ipnet.IP.String(), port)
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
