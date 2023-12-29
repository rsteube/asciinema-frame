package cmd

import (
	"fmt"
	"os"
	"strconv"

	frame "github.com/rsteube/asciinema-frame"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "asciinema-frame [file] [time]",
	Short: "",
	Args:  cobra.ExactArgs(2),
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}

		time, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return err
		}

		f := frame.Frame(file, time)

		switch {
		case cmd.Flag("poster").Changed:
			fmt.Fprintln(cmd.OutOrStdout(), f.Poster())
		default:
			fmt.Fprintln(cmd.OutOrStdout(), f.RawString())
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
func init() {
	rootCmd.Flags().Bool("poster", false, "generate poster string")

	carapace.Gen(rootCmd).PositionalCompletion(
		carapace.ActionFiles(),
	)
}
