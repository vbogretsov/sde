package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"app/pkg/config"
)

var cfg config.Config

var args = cli.Command{
	Commands: []*cli.Command{
		{
			Name:   "svc",
			Usage:  "Start HTTP service",
			Flags:  newSvcFlags(&cfg),
			Action: runSvc,
		},
		{
			Name:   "wrk",
			Usage:  "Start background worker",
			Flags:  newWrkFlags(&cfg),
			Action: runWrk,
		},
		{
			Name: "init",
			Usage: "Initialize application",
			Flags: newInitFlags(&cfg),
			Action: runInit,
		},
	},
}

func main() {
	if err := args.Run(context.Background(), os.Args); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
