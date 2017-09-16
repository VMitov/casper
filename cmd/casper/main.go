package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	casper "github.com/miracl/casper/lib"
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

const (
	defaultPath = "config.yaml"
)

func main() {
	storageFlags := []cli.Flag{
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
	}

	sourcesFlags := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name: "template", Aliases: []string{"t"},
			Usage: "template file",
			Value: "template.yaml",
		}),
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name: "sources", Aliases: []string{"s"},
			Usage: "[key=value, file://file.json]",
			Value: cli.NewStringSlice("file://source.json"),
		}),
	}

	keyFlag := altsrc.NewStringFlag(&cli.StringFlag{
		Name: "key", Aliases: []string{"k"},
		Usage: "specific key to diff",
	})

	app := &cli.App{
		Name:     "casper",
		HelpName: "casper",
		Usage:    "Configuration Automation for Safe and Painless Environment Releases\n" + maskot,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "config", Aliases: []string{"c"},
				Usage: "file to load configurations from",
				Value: defaultPath,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "fetch",
				Aliases: []string{"f"},
				Usage:   "get the current content of a service",
				Flags: append(
					storageFlags,
					altsrc.NewStringFlag(&cli.StringFlag{
						Name: "format", Aliases: []string{"f"},
						Usage: "format of the output",
						Value: "",
					}),
				),
				Action: func(c *cli.Context) error {
					conf, err := newConfig(c.String("config"),
						withTemplate(c.String("template")),
						withSources(c.StringSlice("sources")),
					)
					if err != nil {
						return err
					}

					switch c.String("storage") {
					case "file":
						conf.withFileStorage(c.String("file-path"))
					case "consul":
						if err := conf.withConsulStorage(c.String("consul-addr")); err != nil {
							return err
						}
					default:
						return fmt.Errorf("invalid storage type %v", c.String("storage"))
					}

					out, err := conf.storage.String(c.String("format"))
					if err != nil {
						return err
					}

					fmt.Println(out)
					return nil
				},
			},
			{
				Name:    "build",
				Aliases: []string{"b"},
				Usage:   "build the source for a single service",
				Flags:   sourcesFlags,
				Action: func(c *cli.Context) error {
					conf, err := newConfig(c.String("config"),
						withTemplate(c.String("template")),
						withSources(c.StringSlice("sources")),
					)
					if err != nil {
						return err
					}

					out, err := casper.BuildConfig{
						Tmlp:   conf.template,
						Source: conf.source,
					}.Build()
					if err != nil {
						return err
					}
					fmt.Print(string(out))
					return nil
				},
			},
			{
				Name:    "diff",
				Aliases: []string{"d"},
				Usage:   "show the difference between the source and the content of a service",
				Flags: append([]cli.Flag{
					keyFlag,
					altsrc.NewBoolFlag(&cli.BoolFlag{
						Name: "plain", Aliases: []string{"p"},
						Usage: "disable colorful output",
						Value: false,
					}),
				}, append(storageFlags, sourcesFlags...)...),
				Action: func(c *cli.Context) error {
					conf, err := newConfig(c.String("config"),
						withTemplate(c.String("template")),
						withSources(c.StringSlice("sources")),
					)
					if err != nil {
						return err
					}

					switch c.String("storage") {
					case "file":
						conf.withFileStorage(c.String("file-path"))
					case "consul":
						if err := conf.withConsulStorage(c.String("consul-addr")); err != nil {
							return err
						}
					default:
						return fmt.Errorf("invalid storage type %v", c.String("storage"))
					}

					out, err := casper.BuildConfig{
						Tmlp:   conf.template,
						Source: conf.source,
					}.Build()
					if err != nil {
						return err
					}

					templateName := conf.template.Name()
					templateNameSlice := strings.Split(templateName, ".")
					format := templateNameSlice[len(templateNameSlice)-1]
					// default storage format

					changes, err := conf.storage.GetChanges(out, format, c.String("key"))
					if err != nil {
						return err
					}

					fmt.Println(strChanges(changes, c.String("key"), conf.storage, !c.Bool("plain")))
					return nil
				},
			},
			{
				Name:    "push",
				Aliases: []string{"p"},
				Usage:   "push the source for a service",
				Flags: append([]cli.Flag{
					keyFlag,
					altsrc.NewBoolFlag(&cli.BoolFlag{
						Name:  "force",
						Usage: "push the changes without asking for confirmation",
						Value: false,
					}),
					altsrc.NewBoolFlag(&cli.BoolFlag{
						Name: "plain", Aliases: []string{"p"},
						Usage: "disable colorful output",
						Value: false,
					}),
				}, append(storageFlags, sourcesFlags...)...),
				Action: func(c *cli.Context) error {
					conf, err := newConfig(c.String("config"),
						withTemplate(c.String("template")),
						withSources(c.StringSlice("sources")),
					)
					if err != nil {
						return err
					}

					switch c.String("storage") {
					case "file":
						conf.withFileStorage(c.String("file-path"))
					case "consul":
						if err := conf.withConsulStorage(c.String("consul-addr")); err != nil {
							return err
						}
					default:
						return fmt.Errorf("invalid storage type %v", c.String("storage"))
					}

					out, err := casper.BuildConfig{
						Tmlp:   conf.template,
						Source: conf.source,
					}.Build()
					if err != nil {
						return err
					}

					templateName := conf.template.Name()
					templateNameSlice := strings.Split(templateName, ".")
					format := templateNameSlice[len(templateNameSlice)-1]
					// default storage format

					changes, err := conf.storage.GetChanges(out, format, c.String("key"))
					if err != nil {
						return err
					}

					pretty := !c.Bool("plain")
					fmt.Println(strChanges(changes, c.String("key"), conf.storage, pretty))
					if changes.Len() == 0 {
						return nil
					}

					if !c.Bool("force") {
						// prompt for agreement
						fmt.Print("Continue[y/N]: ")
						input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
						if strings.ToLower(strings.TrimRight(input, "\r\n")) != "y" {
							fmt.Println("Canceled")
							return nil
						}
					}

					fmt.Println("Applying changes...")
					return conf.storage.Push(changes)
				},
			},
		},
	}

	inputSource := func(context *cli.Context) (altsrc.InputSourceContext, error) {
		config := context.String("config")
		_, err := os.Open(config)
		if os.IsNotExist(err) {
			return &altsrc.MapInputSource{}, nil
		}

		return altsrc.NewYamlSourceFromFlagFunc("config")(context)
	}

	// inputSource := altsrc.NewYamlSourceFromFlagFunc("config")
	app.Before = altsrc.InitInputSourceWithContext(app.Flags, inputSource)
	for _, cmd := range app.Commands {
		cmd.Before = altsrc.InitInputSourceWithContext(cmd.Flags, inputSource)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
