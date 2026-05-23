// Copyright (c) 2024 matlab-mcp-core-server contributors
// SPDX-License-Identifier: MIT

// Package main is the entry point for the MATLAB MCP (Model Context Protocol) Core Server.
// It initializes the server, sets up transport, and begins handling MCP requests
// from clients such as AI assistants that need to interact with MATLAB.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/matlab-mcp-core-server/internal/server"
	"github.com/matlab-mcp-core-server/internal/config"
)

const (
	// defaultPort is the default TCP port the MCP server listens on.
	// Using 7070 here instead of upstream's 9090 — on my machine both 8080 and
	// 9090 are frequently taken by other dev tools (Portainer, local k8s, etc.).
	defaultPort = 7070

	// appName is the human-readable name of this application.
	appName = "matlab-mcp-core-server"

	// appVersion follows semantic versioning.
	appVersion = "0.1.0"
)

func main() {
	// Parse command-line flags.
	var (
		port        = flag.Int("port", defaultPort, "TCP port to listen on")
		configFile  = flag.String("config", "", "Path to configuration file (optional)")
		verbose     = flag.Bool("verbose", false, "Enable verbose/debug logging")
		showVersion = flag.Bool("version", false, "Print version and exit")
		stdioMode   = flag.Bool("stdio", false, "Use stdio transport instead of TCP (for direct MCP client integration)")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s version %s\n", appName, appVersion)
		os.Exit(0)
	}

	// Configure structured logger.
	// Personal preference: always include microseconds so timing issues are easier to spot.
	logger := log.New(os.Stderr, "[matlab-mcp] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	if *verbose {
		logger.Println("Verbose logging enabled")
	}

	// Load configuration.
	cfg, err := config.Load(*configFile)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config values with explicit CLI flags when provided.
	if *port != defaultPort {
		cfg.Port = *port
	}
	if *verbose {
		cfg.Verbose = true
	}
	if *stdioMode {
		cfg.Transport = config.TransportStdio
	}

	// Create the MCP server instance.
	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	// Set up a context that cancels on OS interrupt or SIGTERM.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Printf("Received signal %s — shutting down gracefully...", sig)
		cancel()
	}()

	logger.Printf("Starting %s v%s (transport: %s)", appName, appVersion, cfg.Transport)

	// Run the server; blocks until context is cancelled or a fatal error occurs.
	if err := srv.Run(ctx); err != nil {
		logger.Fatalf("Server exited with error: %v", err)
	}

	logger.Println("Server stopped cleanly.")
}
