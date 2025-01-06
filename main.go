package main

import (
	"context"
	"fmt"
	"os"

	"github.com/OffchainLabs/unclog/cmd/check"
	"github.com/OffchainLabs/unclog/cmd/release"
)

var subcommands = map[string]func(context.Context, []string) error{
	"release": release.Run,
	"check":   check.Run,
}

func errExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func printUsage() {
	fmt.Println("Usage: changelog <subcomand> [args]")
	fmt.Println("Subcommands:")
	for sc := range subcommands {
		fmt.Println("\t- ", sc)
	}
}

func main() {
	if err := dispatch(); err != nil {
		errExit(err)
	}
}

func dispatch() error {
	args := os.Args
	if len(args) < 2 {
		printUsage()
		return fmt.Errorf("no command specified")
	}
	sub, args := args[1], args[2:]
	run := subcommands[sub]
	if run == nil {
		printUsage()
		return fmt.Errorf("invalid subcommand %s", sub)
	}
	if err := run(context.Background(), args); err != nil {
		errExit(err)
	}
	return nil
}
