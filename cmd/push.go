package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/miracl/casper/lib/diff"
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

		noValidation, err := cmd.Flags().GetBool("no-validation")
		if err != nil {
			return err
		}

		key, err := cmd.Flags().GetString("key")
		if err != nil {
			return err
		}

		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			return err
		}

		backup, err := cmd.Flags().GetBool("backup")
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

		pushConf := &pushConfig{
			template, noValidation, format, storage, key,
			interactive, backup, &sourcesList, force, !plain,
		}

		return pushRun(pushConf, config)
	},
}

type pushConfig struct {
	template     string
	noValidation bool
	format       string
	storage      string
	key          string
	interactive  bool
	backup       bool
	sourcesList  *[]map[string]interface{}
	force        bool
	pretty       bool
}

func init() {
	pushCmd.Flags().StringP("template", "t", "", "template file")
	pushCmd.Flags().BoolP("no-validation", "n", false, "don't validate input template. Meant for importing from backup")
	pushCmd.Flags().StringP("format", "f", "", "format of the template file")
	pushCmd.Flags().StringP("key", "k", "", "specific key to push, supports a regular expression to match multiple")
	pushCmd.Flags().BoolP("interactive", "i", false, "interactive selection for which keys to push")
	pushCmd.Flags().Bool("force", false, "push the changes without asking")
	pushCmd.Flags().BoolP("plain", "p", false, "disable colorful output")
	pushCmd.Flags().BoolP("backup", "b", false, "back up current configuration to file before applying changes")
	RootCmd.AddCommand(pushCmd)
}

func pushRun(cmdConf *pushConfig, storageConf map[string]interface{}) error {
	out, err := buildConfig(cmdConf.template, !cmdConf.noValidation, *cmdConf.sourcesList)
	if err != nil {
		return err
	}

	s, err := getStorage(cmdConf.storage, storageConf)
	if err != nil {
		return err
	}

	changeset, err := s.GetChanges(out, cmdConf.format, cmdConf.key)
	if err != nil {
		return err
	}

	if changeset.Len() == 0 {
		fmt.Println("No changes")
		return nil
	}

	if cmdConf.interactive && changeset.SupportsInteractive() {
		changeset = filterChangesInteractive(changeset, cmdConf.pretty)
		if changeset.Len() == 0 {
			fmt.Println("No changes have been selected. Nothing to do.")
			return nil
		}
	}

	fmt.Println("\nThe following changes will be applied:")
	fmt.Println(strChanges(changeset, cmdConf.key, s, cmdConf.pretty))

	if !cmdConf.force {
		// prompt for agreement
		if !prompt("Continue [y/N]: ") {
			fmt.Println("Canceled")
			return nil
		}
	}

	// Save a back-up copy of the current configuration
	if cmdConf.backup {
		if err := backup(cmdConf, storageConf); err != nil {
			return nil
		}
	}

	fmt.Println("Applying changes...")
	if err = s.Push(changeset); err != nil {
		return err
	}

	fmt.Println("Done.")
	return nil
}

// filterChangesInteractive prompts the user key-by-key
// if change entries within a set of changes should be applied.
// It then returns the list of changes the user has chosen.
func filterChangesInteractive(changeset changes, pretty bool) changes {
	fmt.Print(
		"Please select the configuration records you would like to apply.\n" +
			"You will be able to confirm the list before applying it.\n\n")

	return changeset.Refine(func(a interface{}) bool {
		b := a.(diff.KVChange)
		var changesDisplay string

		if pretty {
			changesDisplay = b.Pretty()
		} else {
			changesDisplay = b.String()
		}

		return prompt(fmt.Sprintf("\t%s [y/N]: ", changesDisplay))
	}).(changes)
}

// prompt prints out question formatted with params and
// returns true if the user enters "Y" using their keyboard.
func prompt(question string, params ...interface{}) bool {
	fmt.Printf(question, params...)
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	return strings.ToLower(strings.TrimRight(input, "\r\n")) == "y"
}

// backup creates a text file containing a service's current configuration.
func backup(cmdConf *pushConfig, storageConf map[string]interface{}) error {
	currentConf, err := fetchRun(cmdConf.storage, storageConf, cmdConf.format)
	if err != nil {
		return err
	}

	filename, err := saveBackup(currentConf)
	if err != nil {
		return err
	}

	fmt.Printf("Backup has been saved as %s\n", filename)
	return nil
}

// generateBackupFilename generates a filename for a backup file
// where service configuration is to be stored.
// It follows a <UNIX time>_backup.txt format.
func generateBackupFilename() string {
	filenameTemplate := "%d_backup.txt"
	return fmt.Sprintf(filenameTemplate, int32(time.Now().Unix()))
}

// saveBackup generates a text file containing the value of
// its content parameter.
func saveBackup(content *string) (filename string, err error) {
	filename = generateBackupFilename()
	return filename, ioutil.WriteFile(filename, []byte(*content), 0644)
}
