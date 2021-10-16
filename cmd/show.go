package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Decrypt given SealedSecret and print in Secret resource format",
	Long:  `Decrypt given SealedSecret and print in Secret resource format`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("show")
	},
}
