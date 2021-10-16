package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/shusugmt/kubectl-seal/seal"
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
	Short: "Decrypt given SealedSecret and print in Secret resource format",
	Long:  `Decrypt given SealedSecret and print in Secret resource format`,
	Run: func(cmd *cobra.Command, args []string) {

		sealedSecretYAML, err := ioutil.ReadFile(showCmdOpts.filename)
		if err != nil {
			log.Fatalf("%v", err)
		}

		s, err := seal.Unseal(sealedSecretYAML, showCmdOpts.sealedSecretsControllerNamespace)
		if err != nil {
			log.Fatalf("%v", err)
		}

		fmt.Println(string(s))
	},
}
