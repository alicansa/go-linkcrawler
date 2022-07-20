package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/alicansa/go-linkcrawler/crawler"
	"github.com/alicansa/go-linkcrawler/dal/postgres"
	"github.com/alicansa/go-linkcrawler/server"
)

type Main struct {
	// HTTP server for handling HTTP communication.
	HTTPServer *server.Server
	DB         *postgres.DB
}

const (
	host     = "host.docker.internal"
	port     = 5455
	user     = "test_user"
	password = "test_pw"
	dbname   = "linkcrawler_db"
)

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

	if m.DB != nil {
		if err := m.DB.Close(); err != nil {
			return err
		}
	}
	return nil
}

func NewMain() *Main {
	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	return &Main{
		DB: postgres.NewDB(dsn),
	}
}

// Run executes the program. The configuration should already be set up before
// calling this function.
func (m *Main) Run(ctx context.Context) (err error) {

	//Connect to DB
	if err := m.DB.Open(); err != nil {
		return err
	}

	log.Printf("connected to db: dsn=%q", m.DB.DSN)

	//create dal
	linkRepo := postgres.NewLinkRepository(m.DB)
	crawlJobRepo := postgres.NewCrawlJobRepository(m.DB)

	//crawler creator
	createCrawler := func(pe crawler.CrawlPolicyExecuter) crawler.WebCrawler {
		httpClient := &http.Client{}
		return crawler.NewCrawler(httpClient, pe)
	}

	//create handlers
	linksHandler := server.NewLinksHandler(linkRepo)
	crawlJobsHandler := server.NewCrawlJobHandler(crawlJobRepo, linkRepo, createCrawler)

	//create server
	m.HTTPServer = server.NewServer(linksHandler, crawlJobsHandler)

	// Start the HTTP server.
	var port int
	flag.IntVar(&port, "p", 0, "port number")
	flag.Parse()
	m.HTTPServer.Addr = ":" + strconv.Itoa(port)

	if err := m.HTTPServer.Open(); err != nil {
		return err
	}

	log.Printf("running: url=%q", m.HTTPServer.URL())

	return nil
}
