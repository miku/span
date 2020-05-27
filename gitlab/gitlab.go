// Package gitlab contains support types for gitlab interaction.
package gitlab

import (
	"regexp"
)

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
		}
	}
	return
}
