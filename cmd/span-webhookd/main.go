// span-webhookd can serve as a webhook receiver[1] for gitlab, refs #13499
//
// We listen for merge requests[2] to trigger index reviews.
//
// [1] https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#example-webhook-receiver
// [2] https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#merge-request-events
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
	log "github.com/sirupsen/logrus"
)

var (
	addr      = flag.String("addr", ":8080", "hostport to listen on")
	token     = flag.String("token", "", "gitlab auth token, if empty try -token-file")
	tokenFile = flag.String("token-file", path.Join(UserHomeDir(), ".config/span/gitlab.token"), "fallback file for token")
	repoDir   = flag.String("repo-dir", path.Join(os.TempDir(), "span-webhookd/span"), "local repo clone path")
	banner    = `
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
	SolrServer       string
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
		log.Printf("worker received review request: %s", rr)
		log.Println("XXX: running review")

		cmd := "span-review"
		args := []string{"-a", "-server", rr.SolrServer}
		out, err := exec.Command(cmd, args...).Output()

		if err != nil {
			log.Println("%s failed: %s", cmd, err)
			continue
		}

		log.Println(string(out)) // XXX: Post into ticket.
		log.Println("successfully completed review")
	}
	log.Println("worker shutdown")
	done <- true
}

// Repo points to a local copy of the repository containing the configuration
// we want.
type Repo struct {
	URL   string
	Dir   string
	Token string
}

func (r Repo) AuthURL() string {
	return strings.Replace(r.URL, "https://", fmt.Sprintf("https://oauth2:%s@", r.Token), 1)
}

func (r Repo) String() string {
	return fmt.Sprintf("clone from %s at %s", r.URL, r.Dir)
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

// ReadFile reads a file from the repo.
func (r Repo) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(r.Dir, filename))
}

// PushPayload delivered on push and edits with Web IDE.
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
	gitlabEvent := strings.TrimSpace(r.Header.Get("X-Gitlab-Event"))
	switch gitlabEvent {
	case "Push Hook":
		var payload PushPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("gitlab payload: %s", payload)
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

		_, err := repo.ReadFile("docs/review.yaml")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rr := IndexReviewRequest{SolrServer: "dummy", ReviewConfigFile: "sample"}
		IndexReviewQueue <- rr
		log.Println("index review request sent")
	default:
		log.Printf("unregistered or invalid event kind: %s", gitlabEvent)
	}
	log.Printf("request completed after %s", time.Since(started))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintf(w, "This is span-webhookd, a webhook receiver for gitlab.\n"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// UserHomeDir returns the home directory of the user.
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

func main() {
	flag.Parse()

	if *token == "" {
		b, err := ioutil.ReadFile(*tokenFile)
		if err != nil {
			log.Fatal(err)
		}
		*token = strings.TrimSpace(string(b))
	}
	if len(*token) < 10 {
		log.Fatal("auth token too short: %d", len(*token))
	}

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/trigger", HookHandler)
	http.Handle("/", r)

	log.Println(banner)
	log.Printf("starting GitLab webhook receiver on %s ... (settings/integrations)", *addr)

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
