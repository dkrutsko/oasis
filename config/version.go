package config

import (
	"runtime"
	"time"

	"github.com/dkrutsko/oasis/utility"
)

////////////////////////////////////////////////////////////////////////////////

var buildDate string

////////////////////////////////////////////////////////////////////////////////

type Runtime struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

////////////////////////////////////////////////////////////////////////////////

type Version struct {
	Date    time.Time           `json:"date"`
	Build   uint16              `json:"build"`
	Rev     uint16              `json:"rev"`
	Runtime *Runtime            `json:"runtime"`
	Git     *utility.GitDetails `json:"git"`
}

////////////////////////////////////////////////////////////////////////////////

func getDateInfo() (time.Time, uint16, uint16) {

	// If date specified
	if buildDate == "" {
		return time.Time{}, 0, 0
	}

	// Attempt to decode the embedded build date
	t, err := time.Parse(time.RFC3339, buildDate)
	if err != nil {
		return time.Time{}, 0, 0
	}

	// Create reference start date that is January 1, 2000
	base := time.Date(2000, 1, 1, 0, 0, 0, 0, t.Location())

	// Calculate the number of days since base
	build := uint16(t.Sub(base).Hours() / 24)

	// Calculate the number of half seconds since midnight
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	revision := uint16(t.Sub(midnight).Seconds() / 2)

	return t, build, revision
}

////////////////////////////////////////////////////////////////////////////////

func GetVersion() *Version {

	// Attempt to get the date info
	date, build, rev := getDateInfo()

	// Build runtime info
	goRuntime := &Runtime{
		Version: runtime.Version(),
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	// Try to get details from binary
	git, _ := utility.GetGitDetails()

	return &Version{
		Date:    date,
		Build:   build,
		Rev:     rev,
		Runtime: goRuntime,
		Git:     git,
	}
}
