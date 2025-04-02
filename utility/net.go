package utility

import (
	"net"
	"net/http"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

func GetIpFromRequest(req *http.Request, headers bool) (net.IP, error) {

	var addr string

	if headers {
		// Try using real-ip header for ip
		addr = req.Header.Get("x-real-ip")
		if addr == "" {

			// Try using forwarded-for header for ip
			addr = req.Header.Get("x-forwarded-for")
			if addr == "" {

				// Fallback to direct
				addr = req.RemoteAddr
			}
		}

	} else {
		// Don't trust header
		addr = req.RemoteAddr
	}

	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		// Maybe not in an ip:port format
		ip = addr
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, errors.New(
			"ip is not a valid ip",
			errors.String("ip", ip),
		)
	}

	return userIP, nil
}
