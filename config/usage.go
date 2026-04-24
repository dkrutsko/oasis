package config

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////

// Usage prints command-line help for this app.
func Usage() {

	fmt.Print(
		`
SUMMARY
-------
DOCS:

Repository: https://github.com/dkrutsko/oasis

ARGUMENTS
---------
	-version (false) - boolean
	 Print the version and exit.

	-json (false) - boolean
	 Format log output as JSON lines.

	-debug (false) - boolean
	 Whether to output extended logging information for debugging.
`,
	)
}
