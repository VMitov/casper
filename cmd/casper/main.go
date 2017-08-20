package main

import (
	"fmt"
	"os"

	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"
)

const maskot = `
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
		altsrc.NewStringFlag(&cli.StringFlag{
			Name: "template", Aliases: []string{"t"},
			Usage: "template file",
			Value: "template.yaml",
		}),
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name: "sources", Aliases: []string{"s"},
			Usage: "[key=value, file://file.json]",
			Value: cli.NewStringSlice("file://sources.json"),
		}),
	}

	app := &cli.App{
		Name:     "casper",
		HelpName: "casper",
		Usage:    "Configuration Automation for Safe and Painless Environment Releases\n" + maskot,
		Flags: []cli.Flag{
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:  "storage",
				Usage: "[file, consul]",
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:  "file-path",
				Usage: "casper.json",
				Value: "casper.json",
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:  "consul-addr",
				Usage: "http://127.0.0.1:8500/?token=the_one_ring",
				Value: "http://127.0.0.1:8500/",
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name: "ignore", Aliases: []string{"i"},
				Usage: "keys given this value will be ignored by Casper",
				Value: "_ignore",
			}),
			&cli.StringFlag{
				Name:  "config",
				Usage: "file to load configurations from",
				Value: "config.yaml",
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
					return buildRun(c.String("template"), c.StringSlice("sources"))
				},
			},
			{
				Name:    "diff",
				Aliases: []string{"d"},
				Usage:   "show the difference between the source and the content of a service",
				Flags: append([]cli.Flag{
					altsrc.NewStringFlag(&cli.StringFlag{
						Name: "key", Aliases: []string{"k"},
						Usage: "specific key to diff",
					}),
					altsrc.NewStringFlag(&cli.StringFlag{
						Name: "plain", Aliases: []string{"p"},
						Usage: "specific key to diff",
					}),
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
					altsrc.NewStringFlag(&cli.StringFlag{
						Name:  "force",
						Usage: "push the changes without asking for confirmation",
					}),
					altsrc.NewStringFlag(&cli.StringFlag{
						Name: "key", Aliases: []string{"k"},
						Usage: "specific key to diff",
					}),
				}, sourcesFlags...),
				Action: func(c *cli.Context) error {
					fmt.Println("push: ", c.Args())
					return nil
				},
			},
		},
	}

	inputSource := altsrc.NewYamlSourceFromFlagFunc("config")
	app.Before = altsrc.InitInputSourceWithContext(app.Flags, inputSource)
	for _, cmd := range app.Commands {
		cmd.Before = altsrc.InitInputSourceWithContext(cmd.Flags, inputSource)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
