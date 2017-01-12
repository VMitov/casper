package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// buildCmd represents the build command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Get the current configuration of single service",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlag("format", cmd.Flags().Lookup("format"))

		storage := viper.GetString("storage.type")
		config := viper.GetStringMap("storage.config")
		format := viper.GetString("format")

		return fetchRun(storage, config, format)
	},
}

func init() {
	fetchCmd.Flags().StringP("format", "f", "", "format of the output")
	RootCmd.AddCommand(fetchCmd)
}

func fetchRun(storage string, config map[string]interface{}, format string) error {
	s, err := getStorage(storage, config)
	if err != nil {
		return err
	}

	cfg, err := s.String(format)
	if err != nil {
		return err
	}
	fmt.Println(cfg)
	return nil
}
