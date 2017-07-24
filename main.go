package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

const casper = `
	     .-----.
	   .' -   - '.
	  /  .-. .-.  \
	  |  | | | |  |
	   \ \o/ \o/ /
	  _/    ^    \_
	 | \  '---'  / |
	 / /'--. .--'\ \
	/ /'---' '---'\ \
	'.__.       .__.'
	    '|     |'
	     |     \
	     \      '--.
	      '.        '\
	        ''---.   |
	           ,__) /
	            '..'`

func main() {
	sourcesFlags := []cli.Flag{
		&cli.StringFlag{
			Name: "template", Aliases: []string{"t"},
			Usage: "template file",
			Value: "template.yaml",
		},
		&cli.StringSliceFlag{
			Name: "sources", Aliases: []string{"s"},
			Usage: "[key=value, file://file.json]",
			Value: cli.NewStringSlice("file://sources.json"),
		},
	}

	app := &cli.App{
		Name:     "casper",
		HelpName: "casper",
		Usage:    "Configuration Automation for Safe and Painless Environment Releases\n" + casper,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "storage",
				Usage: "[file, consul]",
			},
			&cli.StringFlag{
				Name:  "file-path",
				Usage: "casper.json",
				Value: "casper.json",
			},
			&cli.StringFlag{
				Name:  "consul-addr",
				Usage: "http://127.0.0.1:8500/?token=the_one_ring",
				Value: "http://127.0.0.1:8500/",
			},

			&cli.StringFlag{
				Name: "ignore", Aliases: []string{"i"},
				Usage: "keys given this value will be ignored by Casper",
				Value: "_ignore",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "fetch",
				Aliases: []string{"f"},
				Usage:   "get the current content of a service",
				Action: func(c *cli.Context) error {
					fmt.Println("fetch: ", c.Args(), c.String("storage"), c.LocalFlagNames(), c.FlagNames())
					return nil
				},
			},
			{
				Name:    "build",
				Aliases: []string{"b"},
				Usage:   "build the source for a single service",
				Flags:   sourcesFlags,
				Action: func(c *cli.Context) error {
					fmt.Println("build: ", c.Args())
					return nil
				},
			},
			{
				Name:    "diff",
				Aliases: []string{"d"},
				Usage:   "show the difference between the source and the content of a service",
				Flags: append([]cli.Flag{
					&cli.StringFlag{
						Name: "key", Aliases: []string{"k"},
						Usage: "specific key to diff",
					},
					&cli.StringFlag{
						Name: "plain", Aliases: []string{"p"},
						Usage: "specific key to diff",
					},
				}, sourcesFlags...),
				Action: func(c *cli.Context) error {
					fmt.Println("diff: ", c.Args())
					return nil
				},
			},
			{
				Name:    "push",
				Aliases: []string{"p"},
				Usage:   "push the source for a service",
				Flags: append([]cli.Flag{
					&cli.StringFlag{
						Name:  "force",
						Usage: "push the changes without asking for confirmation",
					},
					&cli.StringFlag{
						Name: "key", Aliases: []string{"k"},
						Usage: "specific key to diff",
					},
				}, sourcesFlags...),
				Action: func(c *cli.Context) error {
					fmt.Println("push: ", c.Args())
					return nil
				},
			},
		},
	}

	app.Run(os.Args)
}
