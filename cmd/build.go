package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	casper "github.com/miracl/casper/lib"
)

var errFormat = errors.New("Sources invalid format")

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the source for a single service",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlag("template", cmd.Flags().Lookup("template"))

		template := viper.GetString("template")
		sourcesList, ok := getSliceStringMapIface(viper.Get("sources"))
		if !ok {
			return errFormat
		}

		return buildRun(template, sourcesList)
	},
}

func init() {
	buildCmd.Flags().StringP("template", "t", "", "template file")
	RootCmd.AddCommand(buildCmd)
}

func buildRun(tmplF string, srcs []map[string]interface{}) error {
	out, err := buildConfig(tmplF, true, srcs)
	if err != nil {
		return err
	}

	fmt.Print(string(out))
	return nil
}

func buildConfig(tmplF string, validate bool, srcs []map[string]interface{}) ([]byte, error) {
	tmpl, err := os.Open(tmplF)
	if err != nil {
		return nil, err
	}

	source, err := getMultiSourcer(srcs)
	if err != nil {
		return nil, err
	}

	var cfg []byte
	config := casper.BuildConfig{
		Tmlp:   tmpl,
		Source: source,
	}

	if validate {
		cfg, err = config.Build()
	} else {
		cfg, err = config.BuildNoValidation()
	}

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
