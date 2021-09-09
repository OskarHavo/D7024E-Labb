package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// exitCmd represents the exit command
var exitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Terminates the node.",
	Long:  `Exit called!`,
	Run: func(cmd *cobra.Command, args []string) {
		exit()
	},
}

func init() {
	rootCmd.AddCommand(exitCmd)
}
func exit() {
	// Terminate node.
	fmt.Println("exit called")

}
