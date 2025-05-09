package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kilicmu/verilis/internal/command"
)

// Variables will be populated at build time
var (
	Version = ""
	AppName = ""
)

func main() {
	helpFlag := flag.Bool("help", false, "Show help information")

	flag.Parse()
	if *helpFlag || len(flag.Args()) == 0 {
		printHelp()
		return
	}

	subCommand := flag.Arg(0)

	executor := command.NewExecutor()

	switch strings.ToLower(subCommand) {
	case "init":
		executor.Init()
	case "generate":
		executor.Generate()
	default:
		fmt.Fprintf(os.Stderr, "\033[31mUnknown command: %s\033[0m\nRun '%s --help' for usage information\n", subCommand, AppName)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("\033[32m 「%s」 is an Ai-driven i18n solution\033[0m \n", AppName)
	fmt.Printf("\033[32m version: %s\033[0m \n", Version)
	fmt.Println("Usage:")
	fmt.Printf("  %s <command> [arguments...]\n\n", AppName)

	fmt.Println("Commands:")
	fmt.Println("init - initial project in this project")
	fmt.Println("generate - generate language resource")
	fmt.Println("")

	fmt.Println("Global options:")
	flag.PrintDefaults()
}
