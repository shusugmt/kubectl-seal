package cmd

import (
	"fmt"
	"log"
	"os"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/shusugmt/kubectl-sealer/sealer"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type newCmdOptions struct {
	filename   string
	name       string
	namespace  string
	secretType string
	scope      string
}

var newCmdOpts = &newCmdOptions{}

func init() {
	addFlagFilename(newCmd, &newCmdOpts.filename, false)

	newCmd.Flags().StringVar(&newCmdOpts.name, "name", "", "name of the base Secret resource")
	newCmd.Flags().StringVar(&newCmdOpts.namespace, "namespace", corev1.NamespaceDefault, "namespace of the base Secret resource")
	newCmd.Flags().StringVar(&newCmdOpts.secretType, "type", string(corev1.SecretTypeOpaque), "type of the base Secret resource")

	defaultScope := ssv1alpha1.DefaultScope
	newCmd.Flags().StringVar(&newCmdOpts.scope, "scope", defaultScope.String(), "set the scope of the sealed secret")
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "create a new SealedSecret",
	Long:  `Create a new SealedSecret.`,
	Run: func(cmd *cobra.Command, args []string) {

		emptySecretYAML, err := generateEmptySecret(newCmdOpts.name, newCmdOpts.namespace, newCmdOpts.secretType, newCmdOpts.scope)
		if err != nil {
			log.Fatalf("%v", err)
		}

		editedSecretYAML, err := sealer.EditWithEditor(emptySecretYAML)
		if err != nil {
			log.Fatalf("%v", err)
		}

		newSealedSecretYAML, err := sealer.Seal(editedSecretYAML, false)
		if err != nil {
			log.Fatalf("%v", err)
		}

		if newCmdOpts.filename != "" {
			f, err := os.Create(newCmdOpts.filename)
			if err != nil {
				log.Fatalf("failed opening file to overwrite with new SealedSecret: %s: %v", f.Name(), err)
			}
			_, err = f.Write(newSealedSecretYAML)
			if err != nil {
				f.Close()
				log.Fatalf("failed writing new SealedSecret: %s: %v", f.Name(), err)
			}
			err = f.Close()
			if err != nil {
				log.Fatalf("failed writing new SealedSecret: %s: %v", f.Name(), err)
			}
		} else {
			fmt.Print(string(newSealedSecretYAML))
		}
	},
}

func generateEmptySecret(name string, namespace string, secretTypeString string, scopeString string) (emptySecretYAML []byte, err error) {

	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "changeme",
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Type: corev1.SecretTypeOpaque,
	}

	if name != "" {
		secret.ObjectMeta.Name = name
	}

	var scope ssv1alpha1.SealingScope
	err = scope.Set(scopeString)
	if err != nil {
		return nil, fmt.Errorf("failed setting scope: \"%s\": %v", scopeString, err)
	}
	ssv1alpha1.UpdateScopeAnnotations(secret.GetObjectMeta().GetAnnotations(), scope)

	secret.Type, err = checkSecretType(secretTypeString)
	if err != nil {
		return nil, err
	}
	fillWithPlaceholders(&secret)

	emptySecretYAML, err = yaml.Marshal(secret)
	if err != nil {
		return nil, fmt.Errorf("error marshalling kubernetes Secret to YAML: %v", err)
	}
	return emptySecretYAML, nil
}

func checkSecretType(secretTypeString string) (corev1.SecretType, error) {
	switch secretTypeString {
	case string(corev1.SecretTypeOpaque):
		return corev1.SecretTypeOpaque, nil
	case string(corev1.SecretTypeServiceAccountToken):
		return corev1.SecretTypeServiceAccountToken, nil
	case string(corev1.SecretTypeDockercfg):
		return corev1.SecretTypeDockercfg, nil
	case string(corev1.SecretTypeDockerConfigJson):
		return corev1.SecretTypeDockerConfigJson, nil
	case string(corev1.SecretTypeBasicAuth):
		return corev1.SecretTypeBasicAuth, nil
	case string(corev1.SecretTypeSSHAuth):
		return corev1.SecretTypeSSHAuth, nil
	case string(corev1.SecretTypeTLS):
		return corev1.SecretTypeTLS, nil
	case string(corev1.SecretTypeBootstrapToken):
		return corev1.SecretTypeBootstrapToken, nil
	default:
		return "", fmt.Errorf("no such type exists for Secret: \"%s\"", secretTypeString)
	}
}

func fillWithPlaceholders(secret *corev1.Secret) {
	secret.StringData = map[string]string{}
	switch secret.Type {
	case corev1.SecretTypeServiceAccountToken:
		secret.StringData[corev1.ServiceAccountTokenKey] = "changeme"

		secret.Annotations = map[string]string{}
		secret.Annotations[corev1.ServiceAccountNameKey] = "changeme"
		secret.Annotations[corev1.ServiceAccountUIDKey] = "changeme"

	case corev1.SecretTypeDockercfg:
		secret.StringData[corev1.DockerConfigKey] = "changeme"

	case corev1.SecretTypeDockerConfigJson:
		secret.StringData[corev1.DockerConfigJsonKey] = "changeme"

	case corev1.SecretTypeBasicAuth:
		secret.StringData[corev1.BasicAuthUsernameKey] = "changeme"
		secret.StringData[corev1.BasicAuthPasswordKey] = "changeme"

	case corev1.SecretTypeSSHAuth:
		secret.StringData[corev1.SSHAuthPrivateKey] = "changeme"

	case corev1.SecretTypeTLS:
		secret.StringData[corev1.TLSCertKey] = "changeme"
		secret.StringData[corev1.TLSPrivateKeyKey] = "changeme"

	default:
		secret.StringData["change"] = "me"
	}
}
