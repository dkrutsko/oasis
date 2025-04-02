package server

import (
	"net/http"
	"path"
	"time"

	"github.com/google/uuid"

	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/utility"
)

////////////////////////////////////////////////////////////////////////////////

func handleRoot() http.HandlerFunc {

	// Load the web app files
	content := getContentFS()

	// Return handler function with ability to access handler
	return func(res http.ResponseWriter, req *http.Request) {

		//----------------------------------------------------------------------------//

		ipStr := ""
		// Attempt to retrieve the proxied IP address from current request
		if netIP, err := utility.GetIpFromRequest(req, true); err == nil {
			// Convert into string
			ipStr = netIP.String()
		}

		// Create the whole path
		fullPath := req.URL.Path
		if req.URL.RawQuery != "" {
			fullPath += "?" + req.URL.RawQuery
		}

		// Create a unique request ID
		unique := uuid.New().String()
		res.Header().Set("X-Request-ID", unique)

		//----------------------------------------------------------------------------//

		var openErr error
		var f http.File
		code := http.StatusNotFound
		file := "404.html"

		// Make sure that the request is either a HEAD or a GET request
		if req.Method == http.MethodHead || req.Method == http.MethodGet {

			filePath := path.Clean(req.URL.Path[1:])
			// Check whether to use the root page
			if filePath == "" || filePath == "." {
				filePath = "index.html"

			} else if path.Ext(filePath) == "" {
				// Append html to bare page paths
				filePath += ".html"
			}

			// Try and open the requested file
			f, openErr = content.Open(filePath)
			if openErr == nil {
				code = http.StatusOK
				file = filePath

			} else {
				// Try opening 404 file instead
				f, openErr = content.Open(file)
			}

		} else {

			// Only allow HEAD and GET methods
			code = http.StatusMethodNotAllowed
			file = "405.html"

			// Try opening 405 file instead
			f, openErr = content.Open(file)
		}

		//----------------------------------------------------------------------------//

		logger.Dbg(
			"request",
			logger.String("ip", ipStr),
			logger.String("uuid", unique),
			logger.Int("code", code),
			logger.String("status", http.StatusText(code)),
			logger.String("method", req.Method),
			logger.String("path", fullPath),
			logger.String("file", file),
			logger.String("user_agent", req.UserAgent()),
			logger.Error("error", openErr),
		)

		//----------------------------------------------------------------------------//

		if openErr != nil {
			res.WriteHeader(http.StatusInternalServerError)

			_, _ = res.Write(
				// Notify client that app is not embedded
				[]byte("web app content is not available"),
			)

			return
		}

		if code != http.StatusOK {
			res.WriteHeader(code)
		}

		http.ServeContent(res, req, file, time.Now(), f)
		_ = f.Close()

		//----------------------------------------------------------------------------//
	}
}
