package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/dkrutsko/oasis/utility"
)

////////////////////////////////////////////////////////////////////////////////

//go:embed content/*
var content embed.FS

////////////////////////////////////////////////////////////////////////////////

func getContentFS() http.FileSystem {

	// Attempt to load the web app files
	f, err := fs.Sub(content, "content")
	if err != nil {
		// If app isn't built
		f = utility.EmptyFS{}
	}

	return http.FS(f)
}
