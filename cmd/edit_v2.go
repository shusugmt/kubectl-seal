package cmd

import (
	"fmt"
	"log"
	"os"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/shusugmt/kubectl-sealer/sealer"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type editv2CmdOptions struct {
	filename                         string
	sealedSecretsControllerNamespace string
	inPlace                          bool
}

var editv2CmdOpts = &editv2CmdOptions{}

func init() {
	addFlagFilename(editv2Cmd, &editv2CmdOpts.filename)
	setSealedSecretsControllerNamespace(&editv2CmdOpts.sealedSecretsControllerNamespace)
	editv2Cmd.Flags().BoolVarP(&editv2CmdOpts.inPlace, "in-place", "i", false, "enable in-place edit")
}

var editv2Cmd = &cobra.Command{
	Use:   "editv2",
	Short: "editv2",
	Long:  `editv2`,
	Run: func(cmd *cobra.Command, args []string) {

		srcSealedSecretYAML, err := os.ReadFile(editv2CmdOpts.filename)
		if err != nil {
			log.Fatalf("%v", err)
		}

		srcSecretYAML, err := sealer.Unseal(srcSealedSecretYAML, editv2CmdOpts.sealedSecretsControllerNamespace)
		if err != nil {
			log.Fatalf("%v", err)
		}

		editedSecretYAML, err := sealer.EditWithEditor(srcSecretYAML)

		if err != nil {
			log.Fatalf("%v", err)
		}
		if editedSecretYAML == nil {
			fmt.Println("no change")
			os.Exit(0)
		}

		updatedSealedSecretYAML, err := updateSealedSecret(srcSealedSecretYAML, srcSecretYAML, editedSecretYAML)
		if err != nil {
			log.Fatalf("%v", err)
		}

		if editv2CmdOpts.inPlace {
			outFile, err := os.Create(editv2CmdOpts.filename)
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

/*
  update SealedSecret with partial update support
*/
func updateSealedSecret(sealedSecretYAML []byte, secretYAML []byte, editedSecretYAML []byte) (updatedSealedSecretYAML []byte, err error) {

	// build struct from yaml
	var sealedSecret ssv1alpha1.SealedSecret
	err = yaml.UnmarshalStrict(sealedSecretYAML, &sealedSecret)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml to SealedSecret: %v", err)
	}

	var secret corev1.Secret
	err = yaml.UnmarshalStrict(secretYAML, &secret)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml to kubernetes Secret: %v", err)
	}

	var editedSecret corev1.Secret
	err = yaml.UnmarshalStrict(editedSecretYAML, &editedSecret)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml to kubernetes Secret: %v", err)
	}

	// ---- ---- ---- ---- ----
	// step.1 check scope change

	// if scope chas been hanged
	if ssv1alpha1.SecretScope(&editedSecret) != ssv1alpha1.SecretScope(&secret) {
		// then need to re-seal entire Secret
		return sealer.Seal(editedSecretYAML, false)
	}

	// ---- ---- ---- ---- ----
	// step.2 check name/namespace change

	// if scope is strict
	if ssv1alpha1.SecretScope(&editedSecret) == ssv1alpha1.StrictScope {
		// and either namespace or name has been changed
		if secret.Namespace != editedSecret.Namespace || secret.Name != editedSecret.Name {
			// then need to re-seal entire Secret
			return sealer.Seal(editedSecretYAML, false)
		}
	}
	// if scope is namespace-wide
	if ssv1alpha1.SecretScope(&editedSecret) == ssv1alpha1.NamespaceWideScope {
		// and namespace has been changed
		if secret.Namespace != editedSecret.Namespace {
			// then need to re-seal entire Secret
			return sealer.Seal(editedSecretYAML, false)
		}
	}

	// ---- ---- ---- ---- ----
	// step.3 ensure metadata update

	// create a copy of edited Secret for generating a skeleton SealdSecret, which contains
	// all data(e.g. metadata.name, ns, labels, annotations, etc) inheriting source Secret
	// EXCEPT actual secret data(spec.data and spec.stringData, which will be mapped to spec.encryptedData).
	editedSecretCopy := editedSecret.DeepCopy()
	// delete secret data
	editedSecretCopy.StringData = nil
	editedSecretCopy.Data = nil

	editedSecretCopyYAML, err := yaml.Marshal(editedSecretCopy)
	if err != nil {
		return nil, err
	}
	// generate skeleton SealedSecret from edited Secret
	// ensuring all metadata is updated
	newSealedSecretYAML, err := sealer.Seal(editedSecretCopyYAML, true)
	if err != nil {
		return nil, err
	}
	// build struct
	var newSealedSecret ssv1alpha1.SealedSecret
	err = yaml.UnmarshalStrict(newSealedSecretYAML, &newSealedSecret)
	if err != nil {
		log.Fatalf("error unmarshalling yaml to SealedSecret: %v", err)
	}
	// copy spec.encryptedData to skeleton
	// we use dump copy `=` here. this is safe because we don't need original sealedSecret anymore
	newSealedSecret.Spec.EncryptedData = sealedSecret.Spec.EncryptedData

	// ---- ---- ---- ---- ----
	// step.4 fill spec.encryptedData keeping unchanged kv pairs left as-is

	// step.4-1
	// add kv pairs those are entirely new
	addedKeys := sealer.GetKeyDiff(editedSecret.StringData, secret.StringData)
	for _, addedKey := range addedKeys {
		// get raw encrypted value
		value := []byte(editedSecret.StringData[addedKey])
		encryptedValue, err := sealer.EncryptRaw(value, editedSecret)
		if err != nil {
			return nil, err
		}
		newSealedSecret.Spec.EncryptedData[addedKey] = string(encryptedValue)
	}

	// step.4-2
	// update kv pairs those values are changed
	updatedKeyVals := sealer.GetUpdatedExisting(editedSecret.StringData, secret.StringData)
	for k, v := range updatedKeyVals {
		// get raw encrypted value
		value := []byte(v)
		encryptedValue, err := sealer.EncryptRaw(value, editedSecret)
		if err != nil {
			return nil, err
		}
		newSealedSecret.Spec.EncryptedData[k] = string(encryptedValue)
	}

	// step.4-3
	// delete kv pairs those are removed
	deletedKeys := sealer.GetKeyDiff(secret.StringData, editedSecret.StringData)
	for _, deletedKey := range deletedKeys {
		delete(newSealedSecret.Spec.EncryptedData, deletedKey)
	}

	// generate YAML
	updatedSealedSecretYAML, err = yaml.Marshal(newSealedSecret)
	if err != nil {
		return nil, err
	}
	return updatedSealedSecretYAML, nil
}
