package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show the difference between the source and a single service",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlag("template", cmd.Flags().Lookup("template"))
		viper.BindPFlag("format", cmd.Flags().Lookup("format"))

		template := viper.GetString("template")
		format := viper.GetString("format")
		storage := viper.GetString("storage.type")
		config := viper.GetStringMap("storage.config")

		key, err := cmd.Flags().GetString("key")
		if err != nil {
			return err
		}

		sourcesList, ok := getSliceStringMapIface(viper.Get("sources"))
		if !ok {
			return errSourceFormat
		}

		plain, err := cmd.Flags().GetBool("plain")
		if err != nil {
			return err
		}

		return diffRun(template, format, key, sourcesList, storage, config, !plain)
	},
}

func init() {
	diffCmd.Flags().StringP("template", "t", "", "template file")
	diffCmd.Flags().StringP("format", "f", "", "format of the template file")
	diffCmd.Flags().StringP("key", "k", "", "specific key to diff")
	diffCmd.Flags().BoolP("plain", "p", false, "disable colorful output")

	RootCmd.AddCommand(diffCmd)
}

func diffRun(tmpl, format, key string, sourcesList []map[string]interface{}, storage string, config map[string]interface{}, pretty bool) error {
	out, err := buildConfig(tmpl, true, sourcesList)
	if err != nil {
		return err
	}

	s, err := getStorage(storage, config)
	if err != nil {
		return err
	}

	// Select format
	if !s.FormatIsValid(format) {
		format = s.DefaultFormat()
	}

	changes, err := s.GetChanges(out, format, key)
	if err != nil {
		return err
	}

	fmt.Println(strChanges(changes, key, s, pretty))
	return nil
}

func strChanges(cs changes, key string, s storage, pretty bool) string {
	if cs.Len() == 0 {
		if key != "" {
			return fmt.Sprintf("No changes for key %v", key)
		}
		return fmt.Sprintf("No changes")
	}

	return s.Diff(cs, pretty)
}
