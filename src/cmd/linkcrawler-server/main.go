package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/alicansa/go-linkcrawler/server"
)

type Main struct {
	// HTTP server for handling HTTP communication.
	HTTPServer *server.Server
}

func main() {
	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	// Instantiate a new type to represent our application.
	// This type lets us shared setup code with our end-to-end tests.
	m := NewMain()

	// Execute program.
	if err := m.Run(ctx); err != nil {
		m.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Wait for CTRL-C.
	<-ctx.Done()

	// Clean up program.
	if err := m.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (m *Main) Close() error {
	if m.HTTPServer != nil {
		if err := m.HTTPServer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func NewMain() *Main {
	return &Main{
		HTTPServer: server.NewServer(),
	}
}

// Run executes the program. The configuration should already be set up before
// calling this function.
func (m *Main) Run(ctx context.Context) (err error) {

	// Start the HTTP server.
	if err := m.HTTPServer.Open(); err != nil {
		return err
	}

	log.Printf("running: url=%q", m.HTTPServer.URL())

	return nil
}
