package utility

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

// In cases where the application is distributed as a binary, these variables
// will be set with `-ldflags` during compilation. They can then be converted to
// git details. For safety, these variables are all base64 encoded.
var (
	gitEmbedCommit  string
	gitEmbedDate    string
	gitEmbedMessage string
	gitEmbedRemote  string
	gitEmbedBranch  string
	gitEmbedStatus  string
)

////////////////////////////////////////////////////////////////////////////////

// GitDetails provides a structured representation of the current state of a
// Git repository. It captures essential details about the latest commit, the
// remote repository, the current branch, and the working tree's status.
type GitDetails struct {
	// Long represents the full hash of the latest commit.
	Long string `json:"long"`

	// Short represents the abbreviated (7 characters) hash of the latest commit.
	Short string `json:"short"`

	// Date represents the timestamp of the latest commit.
	Date time.Time `json:"date"`

	// Message represents the commit message of the latest commit.
	Message string `json:"message"`

	// Remote represents the URL of the remote repository (typically "origin").
	Remote string `json:"remote"`

	// Branch represents the name of the currently checked-out branch.
	Branch string `json:"branch"`

	// Dirty indicates whether there are uncommitted changes in the working tree.
	Dirty bool `json:"dirty"`
}

////////////////////////////////////////////////////////////////////////////////

// GetGitDetails retrieves Git repository information from embedded variables
// set via `-ldflags` during compilation. Used when the application is
// distributed as a binary without access to the Git repository at runtime.
func GetGitDetails() (*GitDetails, error) {

	// If embed data present
	if gitEmbedCommit == "" {
		return nil, errors.New("embedded git data is empty")
	}

	commitBytes, err := base64.StdEncoding.DecodeString(gitEmbedCommit)
	if err != nil {
		return nil, errors.New(
			"failed to decode embedded base64 field",
			errors.String("field", "gitEmbedCommit"),
			errors.Error("error", err),
		)
	}

	dateBytes, err := base64.StdEncoding.DecodeString(gitEmbedDate)
	if err != nil {
		return nil, errors.New(
			"failed to decode embedded base64 field",
			errors.String("field", "gitEmbedDate"),
			errors.Error("error", err),
		)
	}

	messageBytes, err := base64.StdEncoding.DecodeString(gitEmbedMessage)
	if err != nil {
		return nil, errors.New(
			"failed to decode embedded base64 field",
			errors.String("field", "gitEmbedMessage"),
			errors.Error("error", err),
		)
	}

	remoteBytes, err := base64.StdEncoding.DecodeString(gitEmbedRemote)
	if err != nil {
		return nil, errors.New(
			"failed to decode embedded base64 field",
			errors.String("field", "gitEmbedRemote"),
			errors.Error("error", err),
		)
	}

	branchBytes, err := base64.StdEncoding.DecodeString(gitEmbedBranch)
	if err != nil {
		return nil, errors.New(
			"failed to decode embedded base64 field",
			errors.String("field", "gitEmbedBranch"),
			errors.Error("error", err),
		)
	}

	statusBytes, err := base64.StdEncoding.DecodeString(gitEmbedStatus)
	if err != nil {
		return nil, errors.New(
			"failed to decode embedded base64 field",
			errors.String("field", "gitEmbedStatus"),
			errors.Error("error", err),
		)
	}

	// Convert bytes into regular trimmed strings
	commit := strings.TrimSpace(string(commitBytes))
	dateStr := strings.TrimSpace(string(dateBytes))
	message := strings.TrimSpace(string(messageBytes))
	remote := strings.TrimSpace(string(remoteBytes))
	branch := strings.TrimSpace(string(branchBytes))
	status := strings.TrimSpace(string(statusBytes))

	// Convert embedded date string to time object
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, errors.New(
			"failed to parse date field",
			errors.Error("error", err),
		)
	}

	short := commit
	// Avoid any panics
	if len(short) > 7 {
		short = short[:7]
	}

	return &GitDetails{
		Long:    commit,
		Short:   short,
		Date:    date,
		Message: message,
		Remote:  remote,
		Branch:  branch,
		Dirty:   len(status) > 0,
	}, nil
}
