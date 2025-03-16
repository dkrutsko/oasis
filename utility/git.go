package utility

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

var (
	gitEmbedCommit  string
	gitEmbedDate    string
	gitEmbedMessage string
	gitEmbedRemote  string
	gitEmbedBranch  string
	gitEmbedStatus  string
)

////////////////////////////////////////////////////////////////////////////////

type GitDetails struct {
	Long    string    `json:"long"`
	Short   string    `json:"short"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
	Remote  string    `json:"remote"`
	Branch  string    `json:"branch"`
	Dirty   bool      `json:"dirty"`
}

////////////////////////////////////////////////////////////////////////////////

func GetGitDetails() (*GitDetails, error) {

	// If embed data present
	if gitEmbedCommit == "" {
		return nil, errors.New("embedded git data is empty")
	}

	commitBytes, err := base64.StdEncoding.DecodeString(gitEmbedCommit)
	if err != nil {
		return nil, err
	}

	dateBytes, err := base64.StdEncoding.DecodeString(gitEmbedDate)
	if err != nil {
		return nil, err
	}

	messageBytes, err := base64.StdEncoding.DecodeString(gitEmbedMessage)
	if err != nil {
		return nil, err
	}

	remoteBytes, err := base64.StdEncoding.DecodeString(gitEmbedRemote)
	if err != nil {
		return nil, err
	}

	branchBytes, err := base64.StdEncoding.DecodeString(gitEmbedBranch)
	if err != nil {
		return nil, err
	}

	statusBytes, err := base64.StdEncoding.DecodeString(gitEmbedStatus)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &GitDetails{
		Long:    commit,
		Short:   commit[:7],
		Date:    date,
		Message: message,
		Remote:  remote,
		Branch:  branch,
		Dirty:   len(status) > 0,
	}, nil
}
