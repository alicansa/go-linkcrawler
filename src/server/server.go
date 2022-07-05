package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	server *http.Server
	router *mux.Router
	ln     net.Listener

	// Bind address & domain for the server's listener.
	// If domain is specified, server is run on TLS using acme/autocert.
	Addr   string
	Domain string

	LinksService     *LinksService
	CrawlJobsService *CrawlJobsService
}

// ShutdownTimeout is the time given for outstanding requests to finish before shutdown.
const ShutdownTimeout = 600 * time.Second

func NewServer() *Server {
	// configure server and return a pointer to it
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter(),
	}

	// Our router is wrapped by another function handler to perform some
	// middleware-like tasks if necessary
	s.server.Handler = http.HandlerFunc(s.router.ServeHTTP)
	router := s.router.PathPrefix("/api").Subrouter()
	{
		r := router.PathPrefix("/").Subrouter()
		r.Use(enforceJSONHandler)
		s.registerLinksHandler(r)
		s.registerCrawlJobsHandler(r)
	}

	return s
}

// Port returns the TCP port for the running server.
// This is useful in tests where we allocate a random port by using ":0".
func (s *Server) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

// Scheme returns the URL scheme for the server.
func (s *Server) Scheme() string {
	return "http"
}

// URL returns the local base URL of the running server.
func (s *Server) URL() string {
	scheme, port := s.Scheme(), s.Port()

	// Use localhost unless a domain is specified.
	domain := "localhost"
	if s.Domain != "" {
		domain = s.Domain
	}

	// Return without port if using standard ports.
	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s", s.Scheme(), domain)
	}
	return fmt.Sprintf("%s://%s:%d", s.Scheme(), domain, s.Port())
}

// Open validates the server options and begins listening on the bind address.
func (s *Server) Open() (err error) {
	// Open a listener on our bind address.
	if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
		return err
	}

	// Begin serving requests on the listener. We use Serve() instead of
	// ListenAndServe() because it allows us to check for listen errors (such
	// as trying to use an already open port) synchronously.
	go s.server.Serve(s.ln)

	return nil
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func DecodeJSON[T any](rc io.ReadCloser) (T, error) {
	var decoded T
	if err := json.NewDecoder(rc).Decode(&decoded); err != nil {
		return decoded, err
	}

	return decoded, nil
}

func enforceJSONHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if contentType != "" {
			mt, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				http.Error(w, "Malformed Content-Type header", http.StatusBadRequest)
				return
			}

			if mt != "application/json" {
				http.Error(w, "Content-Type header must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
