package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully.",
	Long:  `Put called!`,
	Run: func(cmd *cobra.Command, args []string) {
		test := put("test")
		fmt.Println(test)
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}
func put(uploadfile string) string {
	// Upload data of file downloaded
	// Check if it can be uploaded
	// if so, output the objects hash

	fmt.Println("put called")
	if uploadfile == "test" {
		fmt.Println("test works")
	}

	return "Placeholder Hash : 34873874387473847328743478"

}
