package cmd

import (
	"fmt"
	"os"

	"github.com/shusugmt/kubectl-sealer/sealer"
	"github.com/spf13/cobra"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addFlagFilename(cmd *cobra.Command, storeTo *string, required bool) {
	cmd.Flags().StringVarP(storeTo, "filename", "f", "", "path to SealedSecret resource")
	cmd.MarkFlagFilename("filename")
	if required {
		cmd.MarkFlagRequired("filename")
	}
}

func setSealedSecretsControllerNamespace(storeTo *string) {
	// default to kube-system, consistent with kubeseal
	*storeTo = sealer.GetEnv("SEALED_SECRETS_CONTROLLER_NAMESPACE", "kube-system")
}

var rootCmd = &cobra.Command{
	Use:   "kubectl-sealer",
	Short: "kubectl-sealer",
	Long:  `kubectl-sealer`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(genkeyCmd)
}
