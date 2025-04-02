package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

type Options struct {
	Group *errgroup.Group
	Gctx  context.Context

	Addr string
	Port uint

	Game *game.Game
}

////////////////////////////////////////////////////////////////////////////////

type Server struct {
	options *Options
}

////////////////////////////////////////////////////////////////////////////////

func New(opts *Options) *Server {

	// Create server
	return &Server{
		options: opts,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) Create() error {

	//----------------------------------------------------------------------------//

	// Create address from addr and port
	address := fmt.Sprintf("%s:%d", s.options.Addr, s.options.Port)

	logger.Info(
		"creating a server",
		logger.String("address", address),
	)

	//----------------------------------------------------------------------------//

	// Encapsulate handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot())

	//----------------------------------------------------------------------------//

	server := &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 180 * time.Second,
	}

	//----------------------------------------------------------------------------//

	s.options.Group.Go(func() error {
		logger.Dbg("calling listen and serve")

		err := server.ListenAndServe()
		// Check if error is because server closed
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	s.options.Group.Go(func() error {
		<-s.options.Gctx.Done() // Wait for cancel call
		logger.Dbg("shutting down the server")

		// Don't wait too long for server shutdown
		ctx, cancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer cancel()

		return server.Shutdown(ctx)
	})

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}
