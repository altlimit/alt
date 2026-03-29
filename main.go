package main

import (
	"fmt"
	"os"

	"github.com/altlimit/alt/cmd"
)

var version = "1.0.0-dev"

const usage = `alt — Acquire Latest Tools
A stateless, zero-config CLI distribution proxy for GitHub Releases.

Usage:
  alt <command> [arguments]

Commands:
  install   [-f] user/repo[@tag]  Install a tool from GitHub Releases
  run       user/repo[@tag]       Run a tool without installing it
  update    [user/repo]           Update installed tools to latest version
  list      [user]                Show installed tools
  link      user/repo <alias>     Create a custom command alias
  unlink    <alias>               Remove an alias without uninstalling
  clean     [user/repo|user]      Remove old cached versions
  purge     user/repo|user        Remove a tool completely
  versions  user/repo             List locally cached versions
  which     <command>             Show the binary path for a command

Options:
  --help, -h       Show this help message
  --version, -v    Show version

Environment:
  GITHUB_TOKEN     GitHub personal access token (raises API rate limit)

Examples:
  alt install altlimit/sitegen            Install a tool
  alt install altlimit/taskr@v0.1.7       Install a specific version
  alt run altlimit/sitegen --help         Run without installing
  alt link altlimit/sitegen sg            Create a short alias
  alt unlink sg                           Remove an alias
  alt update                              Update all installed tools
  alt list altlimit                       Show tools by a user
  alt clean                               Remove old cached versions
  alt purge altlimit/altclaw              Completely remove a tool
  alt which altclaw                       Show binary path

Learn more: https://github.com/altlimit/alt
`

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Print(usage)
		os.Exit(0)
	}

	command := args[0]
	commandArgs := args[1:]

	// Handle subcommand help: alt install -h, alt run --help, etc.
	if len(commandArgs) == 1 && (commandArgs[0] == "-h" || commandArgs[0] == "--help") {
		fmt.Print(usage)
		os.Exit(0)
	}

	var err error

	switch command {
	case "install":
		err = cmd.Install(commandArgs)
	case "run":
		err = cmd.Run(commandArgs)
	case "update":
		err = cmd.Update(commandArgs)
	case "list", "ls":
		err = cmd.List(commandArgs)
	case "link":
		err = cmd.Link(commandArgs)
	case "unlink":
		err = cmd.Unlink(commandArgs)
	case "clean":
		err = cmd.Clean(commandArgs)
	case "purge":
		err = cmd.Purge(commandArgs)
	case "versions":
		err = cmd.Versions(commandArgs)
	case "which":
		err = cmd.Which(commandArgs)
	case "--help", "-h", "help":
		fmt.Print(usage)
		os.Exit(0)
	case "--version", "-v", "version":
		fmt.Printf("alt v%s\n", version)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n\nRun 'alt --help' for usage.\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
