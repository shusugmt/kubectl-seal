package cmd

import (
	"fmt"
	"os"

	"github.com/shusugmt/kubectl-seal/seal"
	"github.com/spf13/cobra"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addFlagFilename(cmd *cobra.Command, storeTo *string) {
	cmd.Flags().StringVarP(storeTo, "filename", "f", "", "path to SealedSecret resource")
	cmd.MarkFlagFilename("filename")
	cmd.MarkFlagRequired("filename")
}

func setSealedSecretsControllerNamespace(storeTo *string) {
	// default to kube-system, consistent with kubeseal
	*storeTo = seal.GetEnv("SEALED_SECRETS_CONTROLLER_NAMESPACE", "kube-system")
}

var rootCmd = &cobra.Command{
	Use:   "kubectl-seal",
	Short: "kubectl-seal",
	Long:  `kubectl-seal`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(editv2Cmd)
}
