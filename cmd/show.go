package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/shusugmt/kubectl-sealer/sealer"
	"github.com/spf13/cobra"
)

type showCmdOptions struct {
	filename                         string
	sealedSecretsControllerNamespace string
}

var showCmdOpts = &showCmdOptions{}

func init() {
	addFlagFilename(showCmd, &showCmdOpts.filename)
	setSealedSecretsControllerNamespace(&showCmdOpts.sealedSecretsControllerNamespace)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "decrypt SealedSecret and print in Secret resource format",
	Long:  `Decrypt SealedSecret and print in Secret resource format.`,
	Run: func(cmd *cobra.Command, args []string) {

		sealedSecretYAML, err := os.ReadFile(showCmdOpts.filename)
		if err != nil {
			log.Fatalf("%v", err)
		}

		s, err := sealer.Unseal(sealedSecretYAML, showCmdOpts.sealedSecretsControllerNamespace)
		if err != nil {
			log.Fatalf("%v", err)
		}

		fmt.Println(string(s))
	},
}
