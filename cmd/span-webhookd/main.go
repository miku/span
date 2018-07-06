// span-webhookd can serve as a webhook receiver[1] for gitlab.
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
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var (
	addr = flag.String("addr", ":8080", "hostport to listen on")
)

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
	default:
		log.Println("TODO")
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintf(w, "This is span-webhookd, a webhook receiver for gitlab.\n"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/trigger", MergeRequestHandler)
	http.Handle("/", r)

	log.Printf("starting server on %s", *addr)

	port, err := parsePort(*addr)
	if err != nil {
		log.Fatal(err)
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("GitLab settings/integrations links")
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok {
			log.Printf("http://%s:%d/trigger", ipnet.IP.String(), port)
		}
	}

	log.Fatal(http.ListenAndServe(*addr, r))
}
