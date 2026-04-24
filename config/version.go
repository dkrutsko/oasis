package config

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/dkrutsko/oasis/utility"
)

////////////////////////////////////////////////////////////////////////////////

// AppName represents the main application name constant.
const AppName = "oasis"

////////////////////////////////////////////////////////////////////////////////

// Major and Minor represent application version constants.
const Major uint16 = 0
const Minor uint16 = 0

////////////////////////////////////////////////////////////////////////////////

// buildDate is set at compile time using `-ldflags`.
var buildDate string

////////////////////////////////////////////////////////////////////////////////

// Runtime contains details about the Go runtime environment.
type Runtime struct {
	// Version represents the version of the Go runtime.
	Version string `json:"version"`

	// OS represents the OS the application is running on.
	OS string `json:"os"`

	// Arch represents the system architecture the application is compiled for.
	Arch string `json:"arch"`
}

////////////////////////////////////////////////////////////////////////////////

// Version contains compile-time data about the application.
type Version struct {
	// Name represents the name constant of the application.
	Name string `json:"name"`

	// Date represents the date the application was built.
	Date time.Time `json:"date"`

	// Major and minor represent application version values.
	// Build and Revision encode `buildDate` into integers.
	Major    uint16 `json:"major"`
	Minor    uint16 `json:"minor"`
	Build    uint16 `json:"build"`
	Revision uint16 `json:"revision"`

	// Runtime represents details about the Go runtime environment.
	Runtime *Runtime `json:"runtime"`

	// Git represents the git details of the application.
	Git *utility.GitDetails `json:"git"`
}

////////////////////////////////////////////////////////////////////////////////

// String returns the formatted version string.
func (v *Version) String() string {

	return fmt.Sprintf(
		"%s/%d.%d.%d.%d (%s, %s, %s)",
		v.Name,
		v.Major,
		v.Minor,
		v.Build,
		v.Revision,
		v.Runtime.Version,
		v.Runtime.OS,
		v.Runtime.Arch,
	)
}

////////////////////////////////////////////////////////////////////////////////

var (
	versionInstance     *Version
	versionInstanceLock sync.Mutex
)

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

// GetVersion returns the version information.
func GetVersion() *Version {

	// Ensure synchronization
	versionInstanceLock.Lock()
	defer versionInstanceLock.Unlock()

	// Check if already loaded
	if versionInstance != nil {
		return versionInstance
	}

	// Attempt to retrieve the date info
	date, build, revision := getDateInfo()

	// Create runtime info
	goRuntime := &Runtime{
		Version: runtime.Version(),
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	// Try to get details from binary
	git, _ := utility.GetGitDetails()

	versionInstance = &Version{
		Name:     AppName,
		Date:     date,
		Major:    Major,
		Minor:    Minor,
		Build:    build,
		Revision: revision,
		Runtime:  goRuntime,
		Git:      git,
	}

	return versionInstance
}
