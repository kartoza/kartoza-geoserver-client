// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/webserver"
)

var (
	version = "dev"
)

func main() {
	// Parse command line flags
	addr := flag.String("addr", ":8080", "HTTP server address")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("kartoza-cloudbench %s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and start web server
	server := webserver.New(cfg)

	fmt.Printf("Starting Kartoza CloudBench %s\n", version)
	fmt.Printf("Server listening on http://localhost%s\n", *addr)
	fmt.Println("Press Ctrl+C to stop")

	if err := server.Start(*addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
