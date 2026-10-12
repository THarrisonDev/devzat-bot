package main

import (
	"fmt"
	"os"
	"strings"

	api "github.com/quackduck/devzat/devzatapi"
)

const (
	welcomeName = "\x1b[1;32mwelcome\x1b[0m"
	remindName  = "\x1b[1;33mremind\x1b[0m"
)

const usage = `Usage:
  devzat-bot <host:port>       Connect and serve the welcome + remind features

Configuration (environment):
  DEVZAT_ADDRESS    Server plugin-API address, e.g. localhost:5556 (arg overrides)
  DEVZAT_TOKEN      Plugin API token
  WELCOME_ENABLED   DM new joiners a guide (default true)
  REMIND_ENABLED    Enable the 'remind' command (default true)
  WELCOME_FILE      Path to a file with custom welcome text ({{name}} substituted)
`

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func enabled(key string) bool {
	return !strings.EqualFold(os.Getenv(key), "false")
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(usage)
		return
	}

	address := os.Getenv("DEVZAT_ADDRESS")
	token := os.Getenv("DEVZAT_TOKEN")
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		address = args[0]
	}
	if address == "" {
		fmt.Fprintln(os.Stderr, "no server address: pass it as the first argument or set DEVZAT_ADDRESS")
		os.Exit(1)
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "no token: set DEVZAT_TOKEN")
		os.Exit(1)
	}

	s, err := api.NewSession(address, token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating session:", err)
		os.Exit(1)
	}
	defer s.Close()

	if enabled("WELCOME_ENABLED") {
		if err := startWelcome(s); err != nil {
			fmt.Fprintln(os.Stderr, "error starting welcome:", err)
			os.Exit(1)
		}
		fmt.Println("welcome enabled")
	}
	if enabled("REMIND_ENABLED") {
		if err := startRemind(s); err != nil {
			fmt.Fprintln(os.Stderr, "error starting remind:", err)
			os.Exit(1)
		}
		fmt.Println("remind enabled")
	}

	fmt.Printf("devzat-bot connected to %s\n", address)
	if err := <-s.ErrorChan; err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
