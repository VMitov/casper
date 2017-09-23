package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/miracl/casper"
	"github.com/pkg/errors"
	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"
)

var version = "devel"

const mascot = `
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
	configFlag       = "config"
	defaultPath      = "config.yaml"
	defaultIgnoreVal = "_ignore"
)

func newApp() *cli.App {
	storageFlags := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "storage",
			Usage:   "[file, consul]",
			Value:   "file",
			EnvVars: []string{"CASPER_STORAGE"},
		}),
		newPathFlag(&cli.StringFlag{
			Name:    "file-path",
			Usage:   "casper.yaml",
			Value:   "casper.yaml",
			EnvVars: []string{"CASPER_FILE_PATH"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "consul-addr",
			Usage:   fmt.Sprintf("http://127.0.0.1:8500/?ignore=%v&token=aclToken", defaultIgnoreVal),
			Value:   fmt.Sprintf("http://127.0.0.1:8500/?ignore=%v", defaultIgnoreVal),
			EnvVars: []string{"CASPER_CONSUL_ADDR"},
		}),
	}

	formatFlag := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name: "format", Aliases: []string{"f"},
			Usage: "format of the output",
			Value: "",
		}),
	}

	sourcesFlags := []cli.Flag{
		newPathFlag(&cli.StringFlag{
			Name: "template", Aliases: []string{"t"},
			Usage:   "template file",
			Value:   "template.yaml",
			EnvVars: []string{"CASPER_TEMPLATE"},
		}),
		newSourcesSliceFlag(&cli.StringSliceFlag{
			Name: "sources", Aliases: []string{"s"},
			Usage:   "[key=value, file://file.yaml]",
			Value:   cli.NewStringSlice("file://source.yaml"),
			EnvVars: []string{"CASPER_SOURCES"},
		}),
	}

	keyFlag := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name: "key", Aliases: []string{"k"},
			Usage: "specific key to diff",
		}),
	}

	plainFlag := []cli.Flag{
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name: "plain", Aliases: []string{"p"},
			Usage:   "disable colorful output",
			Value:   false,
			EnvVars: []string{"CASPER_PLAIN"},
		}),
	}

	forceFlag := []cli.Flag{
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:    "force",
			Usage:   "push the changes without asking for confirmation",
			Value:   false,
			EnvVars: []string{"CASPER_SILENT"},
		}),
	}

	app := &cli.App{
		Name:     "casper",
		HelpName: "casper",
		Version:  version,
		Usage:    "Configuration Automation for Safe and Painless Environment Releases\n" + mascot,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: configFlag, Aliases: []string{"c"},
				Usage:   "file to load configurations from",
				Value:   defaultPath,
				EnvVars: []string{"CASPER_CONFIG"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "fetch",
				Aliases: []string{"f"},
				Usage:   "get the current content of a service",
				Flags:   combineFlags(storageFlags, formatFlag),
				Action:  fetchAction,
			},
			{
				Name:    "build",
				Aliases: []string{"b"},
				Usage:   "build the source for a single service",
				Flags:   sourcesFlags,
				Action:  buildAction,
			},
			{
				Name:    "diff",
				Aliases: []string{"d"},
				Usage:   "show the difference between the source and the content of a service",
				Flags:   combineFlags(storageFlags, sourcesFlags, keyFlag, plainFlag),
				Action:  diffAction,
			},
			{
				Name:    "push",
				Aliases: []string{"p"},
				Usage:   "push the source for a service",
				Flags:   combineFlags(storageFlags, sourcesFlags, keyFlag, plainFlag, forceFlag),
				Action:  pushAction,
			},
		},
	}

	// returns empty altsrc.MapInputSource if config doesn't exists.
	inputSource := func(context *cli.Context) (altsrc.InputSourceContext, error) {
		config := context.String(configFlag)
		_, err := os.Open(config)
		if os.IsNotExist(err) {
			return &altsrc.MapInputSource{}, nil
		}

		return altsrc.NewYamlSourceFromFlagFunc(configFlag)(context)
	}

	app.Before = altsrc.InitInputSourceWithContext(app.Flags, inputSource)
	for _, cmd := range app.Commands {
		cmd.Before = altsrc.InitInputSourceWithContext(cmd.Flags, inputSource)
	}

	return app
}

func main() {
	app := newApp()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func fetchAction(c *cli.Context) error {
	ctx, err := newContext(c.String(configFlag),
		withSources(c.StringSlice("sources")),
	)
	if err != nil {
		return errors.Wrap(err, "creating context failed")
	}

	if err := withStorage(ctx, c); err != nil {
		return err
	}

	out, err := ctx.storage.String(c.String("format"))
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}

func buildAction(c *cli.Context) error {
	ctx, err := newContext(c.String(configFlag),
		withTemplate(c.String("template")),
		withSources(c.StringSlice("sources")),
	)
	if err != nil {
		return errors.Wrap(err, "creating context failed")
	}

	out, err := casper.BuildConfig{
		Template: ctx.template,
		Source:   ctx.source,
	}.Build()
	if err != nil {
		return errors.Wrap(err, "building the source failed")
	}
	fmt.Print(string(out))
	return nil
}

func diffAction(c *cli.Context) error {
	ctx, err := newContext(c.String(configFlag),
		withTemplate(c.String("template")),
		withSources(c.StringSlice("sources")),
	)
	if err != nil {
		return errors.Wrap(err, "creating context failed")
	}

	if err := withStorage(ctx, c); err != nil {
		return err
	}

	out, err := casper.BuildConfig{
		Template: ctx.template,
		Source:   ctx.source,
	}.Build()
	if err != nil {
		return errors.Wrap(err, "building the source failed")
	}

	changes, err := ctx.storage.GetChanges(out, getFormat(ctx), c.String("key"))
	if err != nil {
		return errors.Wrap(err, "getting changes failed")
	}

	fmt.Println(strChanges(changes, c.String("key"), ctx.storage, !c.Bool("plain")))
	return nil
}

func pushAction(c *cli.Context) error {
	ctx, err := newContext(c.String(configFlag),
		withTemplate(c.String("template")),
		withSources(c.StringSlice("sources")),
	)
	if err != nil {
		return errors.Wrap(err, "creating context failed")
	}

	if err := withStorage(ctx, c); err != nil {
		return err
	}

	out, err := casper.BuildConfig{
		Template: ctx.template,
		Source:   ctx.source,
	}.Build()
	if err != nil {
		return errors.Wrap(err, "building the source failed")
	}

	changes, err := ctx.storage.GetChanges(out, getFormat(ctx), c.String("key"))
	if err != nil {
		return errors.Wrap(err, "getting changes failed")
	}

	fmt.Println(strChanges(changes, c.String("key"), ctx.storage, !c.Bool("plain")))
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
	return ctx.storage.Push(changes)
}

func combineFlags(flagLists ...[]cli.Flag) []cli.Flag {
	flags := []cli.Flag{}

	for _, l := range flagLists {
		flags = append(flags, l...)
	}

	return flags
}

func withStorage(ctx *context, c *cli.Context) error {
	switch c.String("storage") {
	case "file":
		ctx.withFileStorage(c.String("file-path"))
	case "consul":
		if err := ctx.withConsulStorage(c.String("consul-addr")); err != nil {
			return errors.Wrap(err, "setting Consul storage failed")
		}
	default:
		return fmt.Errorf("invalid storage type '%v'", c.String("storage"))
	}

	return nil
}

func getFormat(ctx *context) string {
	templateName := ctx.template.Name()
	templateNameSlice := strings.Split(templateName, ".")
	return templateNameSlice[len(templateNameSlice)-1]
}

func strChanges(cs casper.Changes, key string, s casper.Storage, pretty bool) string {
	if cs.Len() == 0 {
		if key != "" {
			return fmt.Sprintf("No changes for key %v", key)
		}
		return fmt.Sprintf("No changes")
	}
	return s.Diff(cs, pretty)
}
