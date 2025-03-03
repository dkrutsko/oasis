package config

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////

func Usage() {

	fmt.Println(`
SUMMARY
-------
DOCS:

Repository: https://github.com/dkrutsko/oasis

ARGUMENTS
---------
  -debug (false) - bool
   DOCS:

  -json (false) - bool
   DOCS:

  -version (false) - bool
   DOCS:

  -addr ("localhost") - string
   DOCS:

  -port (8080) - uint
   DOCS:
`)
}
