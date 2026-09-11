package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const usageText = `vanity - Go vanity import path server

Usage:
  vanity help
  vanity validate -config vanity.yaml
  vanity serve -config vanity.yaml -listen 0.0.0.0:3030

Commands:
  help       Show this help message.
  validate   Validate configuration.
  serve      Start HTTP server.

Options:
  -config PATH
        Path to YAML configuration file.
  -listen ADDRESS
        HTTP listen address (default ":3030").
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 2
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return 0

	case "validate":
		return validate(args[1:])

	case "serve":
		return serve(args[1:])

	default:
		fmt.Fprintf(os.Stderr, "vanity: unknown command %q\n\n", args[0])
		return 2
	}
}

func validate(args []string) int {
	fs := flag.NewFlagSet("vanity validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configFile := fs.String("config", "vanity.yaml", "path to YAML configuration file")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "vanity validate: unexpected arguments")
		return 2
	}

	if _, err := LoadConfig(*configFile); err != nil {
		fmt.Fprintf(os.Stderr, "vanity validate: %v\n", err)
		return 1
	}

	fmt.Printf("valid: %s\n", *configFile)

	return 0
}

func serve(args []string) int {
	fs := flag.NewFlagSet("vanity serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configFile := fs.String(
		"config",
		"vanity.yaml",
		"path to YAML configuration file",
	)
	listenAddr := fs.String(
		"listen",
		":3030",
		"HTTP listen address",
	)

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "vanity serve: unexpected arguments")
		return 2
	}

	config, err := LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vanity serve: %v\n", err)
		return 1
	}

	server, err := NewServer(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vanity serve: %v\n", err)
		return 1
	}

	httpServer := &http.Server{
		Addr:    *listenAddr,
		Handler: server,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	go func() {
		<-stop

		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "vanity serve: shutdown: %v\n", err)
		}
	}()

	fmt.Printf("listening on %s\n", *listenAddr)

	if err := httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "vanity serve: %v\n", err)
		return 1
	}

	return 0
}

func printUsage() {
	fmt.Print(usageText)
}
