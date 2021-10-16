package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/shusugmt/kubectl-seal/seal"
	"github.com/spf13/cobra"
)

type editCmdOptions struct {
	filename                         string
	sealedSecretsControllerNamespace string
	inPlace                          bool
}

var editCmdOpts = &editCmdOptions{}

func init() {
	addFlagFilename(editCmd, &editCmdOpts.filename)
	setSealedSecretsControllerNamespace(&editCmdOpts.sealedSecretsControllerNamespace)
	editCmd.Flags().BoolVarP(&editCmdOpts.inPlace, "in-place", "i", false, "enable in-place edit")
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit",
	Long:  `edit`,
	Run: func(cmd *cobra.Command, args []string) {

		srcSealedSecretYAML, err := ioutil.ReadFile(editCmdOpts.filename)
		if err != nil {
			log.Fatalf("%v", err)
		}

		srcSecretYAML, err := seal.Unseal(srcSealedSecretYAML, "sealed-secrets")
		if err != nil {
			log.Fatalf("%v", err)
		}

		editedSecretYAML, err := seal.EditWithEditor(srcSecretYAML)
		if err != nil {
			log.Fatalf("%v", err)
		}
		if editedSecretYAML == nil {
			fmt.Println("no change")
			os.Exit(0)
		}

		updatedSealedSecretYAML, err := seal.Seal(editedSecretYAML, false)
		if err != nil {
			log.Fatalf("%v", err)
		}

		if editCmdOpts.inPlace {
			outFile, err := os.Create(editCmdOpts.filename)
			if err != nil {
				log.Fatalf("failed opening file to overwrite with updated SealedSecret: %s: %v", outFile.Name(), err)
			}
			_, err = outFile.Write(updatedSealedSecretYAML)
			if err != nil {
				log.Fatalf("failed writing updated SealedSecret: %s: %v", outFile.Name(), err)
			}
		} else {
			fmt.Print(string(updatedSealedSecretYAML))
		}
	},
}
