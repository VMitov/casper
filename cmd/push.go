package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push the source for a single service",
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

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		plain, err := cmd.Flags().GetBool("plain")
		if err != nil {
			return err
		}

		// fix the template path if it is from the config file
		if viper.InConfig("template") && !cmd.Flag("template").Changed {
			template = configPath(cfgFile, template)
		}

		return pushRun(template, format, key, sourcesList, storage, config, force, !plain)
	},
}

func init() {
	pushCmd.Flags().StringP("template", "t", "", "template file")
	pushCmd.Flags().StringP("format", "f", "", "format of the template file")
	pushCmd.Flags().StringP("key", "k", "", "specific key to push")
	pushCmd.Flags().Bool("force", false, "push the changes without asking")
	pushCmd.Flags().BoolP("plain", "p", false, "disable colorful output")
	RootCmd.AddCommand(pushCmd)
}

func pushRun(tmpl, format, key string, sourcesList []map[string]interface{}, storage string, config map[string]interface{}, force, pretty bool) error {
	out, err := buildConfig(tmpl, sourcesList)
	if err != nil {
		return err
	}

	s, err := getStorage(storage, config)
	if err != nil {
		return err
	}

	changes, err := s.GetChanges(out, format, key)
	if err != nil {
		return err
	}

	fmt.Println(strChanges(changes, key, s, pretty))
	if changes.Len() == 0 {
		return nil
	}

	if !force {
		// prompt for agreement
		fmt.Print("Continue[y/N]: ")
		input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.ToLower(strings.TrimRight(input, "\r\n")) != "y" {
			fmt.Println("Canceled")
			return nil
		}
	}

	fmt.Println("Applying changes...")
	return s.Push(changes)
}
