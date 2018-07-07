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
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	addr      = flag.String("addr", ":8080", "hostport to listen on")
	token     = flag.String("token", "", "gitlab auth token, if empty try -token-file")
	tokenFile = flag.String("token-file", path.Join(UserHomeDir(), ".config/span/gitlab.token"), "fallback file, if token is missing")
	repoDir   = flag.String("repo-dir", path.Join(os.TempDir(), "span-webhookd/span"), "local repo clone")
	repoURL   = flag.String("repo-url", "https://git.sc.uni-leipzig.de/miku/span.git", "remote git clone URL")
	banner    = `
                         888       888                        888   _         888
Y88b    e    /  e88~~8e  888-~88e  888-~88e  e88~-_   e88~-_  888 e~ ~   e88~\888
 Y88b  d8b  /  d888  88b 888  888b 888  888 d888   i d888   i 888d8b    d888  888
  Y888/Y88b/   8888__888 888  8888 888  888 8888   | 8888   | 888Y88b   8888  888
   Y8/  Y8/    Y888    , 888  888P 888  888 Y888   ' Y888   ' 888 Y88b  Y888  888
    Y    Y      "88___/  888-_88"  888  888  "88_-~   "88_-~  888  Y88b  "88_/888
`
)

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
	return nil, nil
}

// MergeRequestPayload is sent by gitlab on merge request events.
type MergeRequestPayload struct {
	Changes struct {
		Labels struct {
			Current []struct {
				Color       string `json:"color"`
				CreatedAt   string `json:"created_at"`
				Description string `json:"description"`
				GroupId     int64  `json:"group_id"`
				Id          int64  `json:"id"`
				ProjectId   int64  `json:"project_id"`
				Template    bool   `json:"template"`
				Title       string `json:"title"`
				Type        string `json:"type"`
				UpdatedAt   string `json:"updated_at"`
			} `json:"current"`
			Previous []struct {
				Color       string `json:"color"`
				CreatedAt   string `json:"created_at"`
				Description string `json:"description"`
				GroupId     int64  `json:"group_id"`
				Id          int64  `json:"id"`
				ProjectId   int64  `json:"project_id"`
				Template    bool   `json:"template"`
				Title       string `json:"title"`
				Type        string `json:"type"`
				UpdatedAt   string `json:"updated_at"`
			} `json:"previous"`
		} `json:"labels"`
		UpdatedAt   []string      `json:"updated_at"`
		UpdatedById []interface{} `json:"updated_by_id"`
	} `json:"changes"`
	Labels []struct {
		Color       string `json:"color"`
		CreatedAt   string `json:"created_at"`
		Description string `json:"description"`
		GroupId     int64  `json:"group_id"`
		Id          int64  `json:"id"`
		ProjectId   int64  `json:"project_id"`
		Template    bool   `json:"template"`
		Title       string `json:"title"`
		Type        string `json:"type"`
		UpdatedAt   string `json:"updated_at"`
	} `json:"labels"`
	ObjectAttributes struct {
		Action   string `json:"action"`
		Assignee struct {
			AvatarUrl string `json:"avatar_url"`
			Name      string `json:"name"`
			Username  string `json:"username"`
		} `json:"assignee"`
		AssigneeId  int64  `json:"assignee_id"`
		AuthorId    int64  `json:"author_id"`
		CreatedAt   string `json:"created_at"`
		Description string `json:"description"`
		Id          int64  `json:"id"`
		Iid         int64  `json:"iid"`
		LastCommit  struct {
			Author struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			} `json:"author"`
			Id        string `json:"id"`
			Message   string `json:"message"`
			Timestamp string `json:"timestamp"`
			Url       string `json:"url"`
		} `json:"last_commit"`
		MergeStatus string      `json:"merge_status"`
		MilestoneId interface{} `json:"milestone_id"`
		Source      struct {
			AvatarUrl         interface{} `json:"avatar_url"`
			DefaultBranch     string      `json:"default_branch"`
			Description       string      `json:"description"`
			GitHttpUrl        string      `json:"git_http_url"`
			GitSshUrl         string      `json:"git_ssh_url"`
			Homepage          string      `json:"homepage"`
			HttpUrl           string      `json:"http_url"`
			Name              string      `json:"name"`
			Namespace         string      `json:"namespace"`
			PathWithNamespace string      `json:"path_with_namespace"`
			SshUrl            string      `json:"ssh_url"`
			Url               string      `json:"url"`
			VisibilityLevel   int64       `json:"visibility_level"`
			WebUrl            string      `json:"web_url"`
		} `json:"source"`
		SourceBranch    string `json:"source_branch"`
		SourceProjectId int64  `json:"source_project_id"`
		State           string `json:"state"`
		Target          struct {
			AvatarUrl         interface{} `json:"avatar_url"`
			DefaultBranch     string      `json:"default_branch"`
			Description       string      `json:"description"`
			GitHttpUrl        string      `json:"git_http_url"`
			GitSshUrl         string      `json:"git_ssh_url"`
			Homepage          string      `json:"homepage"`
			HttpUrl           string      `json:"http_url"`
			Name              string      `json:"name"`
			Namespace         string      `json:"namespace"`
			PathWithNamespace string      `json:"path_with_namespace"`
			SshUrl            string      `json:"ssh_url"`
			Url               string      `json:"url"`
			VisibilityLevel   int64       `json:"visibility_level"`
			WebUrl            string      `json:"web_url"`
		} `json:"target"`
		TargetBranch    string `json:"target_branch"`
		TargetProjectId int64  `json:"target_project_id"`
		Title           string `json:"title"`
		UpdatedAt       string `json:"updated_at"`
		Url             string `json:"url"`
		WorkInProgress  bool   `json:"work_in_progress"`
	} `json:"object_attributes"`
	ObjectKind string `json:"object_kind"`
	Project    struct {
		AvatarUrl         interface{} `json:"avatar_url"`
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
	Repository struct {
		Description string `json:"description"`
		Homepage    string `json:"homepage"`
		Name        string `json:"name"`
		Url         string `json:"url"`
	} `json:"repository"`
	User struct {
		AvatarUrl string `json:"avatar_url"`
		Name      string `json:"name"`
		Username  string `json:"username"`
	} `json:"user"`
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

// Example push payload.
//
// {
//   "object_kind": "push",
//   "event_name": "push",
//   "before": "f5fcd387688fe62cbbd6952ff8e2e8d539f16da6",
//   "after": "4f2274ab095010b68d0debbee27213ef12f489a1",
//   "ref": "refs/heads/master",
//   "checkout_sha": "4f2274ab095010b68d0debbee27213ef12f489a1",
//   "message": null,
//   "user_id": 15,
//   "user_name": "Martin Czygan",
//   "user_username": "miku",
//   "user_email": "martin.czygan@uni-leipzig.de",
//   "user_avatar": "https://git.sc.uni-leipzig.de/uploads/-/system/user/avatar/15/avatar.png",
//   "project_id": 46,
//   "project": {
//     "id": 46,
//     "name": "span",
//     "description": "Mirror of span.",
//     "web_url": "https://git.sc.uni-leipzig.de/miku/span",
//     "avatar_url": null,
//     "git_ssh_url": "git@git.sc.uni-leipzig.de:miku/span.git",
//     "git_http_url": "https://git.sc.uni-leipzig.de/miku/span.git",
//     "namespace": "miku",
//     "visibility_level": 10,
//     "path_with_namespace": "miku/span",
//     "default_branch": "master",
//     "ci_config_path": null,
//     "homepage": "https://git.sc.uni-leipzig.de/miku/span",
//     "url": "git@git.sc.uni-leipzig.de:miku/span.git",
//     "ssh_url": "git@git.sc.uni-leipzig.de:miku/span.git",
//     "http_url": "https://git.sc.uni-leipzig.de/miku/span.git"
//   },
//   "commits": [
//     {
//       "id": "4f2274ab095010b68d0debbee27213ef12f489a1",
//       "message": "Update main.go",
//       "timestamp": "2018-07-06T12:18:36+02:00",
//       "url": "https://git.sc.uni-leipzig.de/miku/span/commit/4f2274ab095010b68d0debbee27213ef12f489a1",
//       "author": {
//         "name": "Martin Czygan",
//         "email": "martin.czygan@uni-leipzig.de"
//       },
//       "added": [],
//       "modified": [
//         "cmd/span-webhookd/main.go"
//       ],
//       "removed": []
//     }
//   ],
//   "total_commits_count": 1,
//   "repository": {
//     "name": "span",
//     "url": "git@git.sc.uni-leipzig.de:miku/span.git",
//     "description": "Mirror of span.",
//     "homepage": "https://git.sc.uni-leipzig.de/miku/span",
//     "git_http_url": "https://git.sc.uni-leipzig.de/miku/span.git",
//     "git_ssh_url": "git@git.sc.uni-leipzig.de:miku/span.git",
//     "visibility_level": 10
//   }
// }

func MergeRequestHandler(w http.ResponseWriter, r *http.Request) {
	known := map[string]bool{
		"Push Hook":     true, // Push hook.
		"Issue Hook":    true, // Issue hook.
		"Note Hook":     true, // Comment, issue, comment on code, merge hook.
		"Tag Push Hook": true, // Tag push hook.
	}
	kind := strings.TrimSpace(r.Header.Get("X-Gitlab-Event"))
	if _, ok := known[kind]; !ok {
		log.Printf("unknown event type: %s", kind)
	}
	switch {
	case kind == "Note Hook":
		var payload MergeRequestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case kind == "Push Hook":
		var payload PushPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Println(payload)
		repo := Repo{URL: *repoURL, Dir: *repoDir, Token: *token}
		if err := repo.Update(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Printf("successfully updated repo at %s", repo.Dir)
		// XXX: Update repo, show changed file.
	default:
		log.Printf("TODO (kind=%s)", kind)
	}
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
	r.HandleFunc("/trigger", MergeRequestHandler)
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

	log.Fatal(http.ListenAndServe(*addr, r))
}
