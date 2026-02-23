// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/kartoza/kartoza-cloudbench/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
