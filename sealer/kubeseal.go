package sealer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func Unseal(sealedSecretYAML []byte, sealedSecretsControllerNamespace string) (secretYAML []byte, err error) {
	// create and open a temporary file
	f, err := os.CreateTemp("", "kubectl-sealer-")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file: %v", err)
	}

	// close and remove the temporary file at the end of the program
	defer f.Close()
	defer os.Remove(f.Name())

	// get sealing keys
	kubectlCommandArgs := []string{
		"get", "secret",
		"-l", "sealedsecrets.bitnami.com/sealed-secrets-key",
		"-n", sealedSecretsControllerNamespace,
		"-o", "yaml",
	}
	kubectlCommand := exec.Command("kubectl", kubectlCommandArgs...)
	sealingKeysYAML, err := kubectlCommand.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error invoking kubectl as %v: %v: %s", kubectlCommand.Args, err, sealingKeysYAML)
	}

	// store sealing keys into temporary file
	_, err = f.Write(sealingKeysYAML)
	if err != nil {
		return nil, fmt.Errorf("error writing sealing keys to temporary file: %v", err)
	}

	// unseal
	kubesealCommandArgs := []string{
		"--recovery-unseal",
		"--recovery-private-key", f.Name(),
	}
	kubesealCommand := exec.Command("kubeseal", kubesealCommandArgs...)
	kubesealCommandStdin, _ := kubesealCommand.StdinPipe()
	_, err = kubesealCommandStdin.Write(sealedSecretYAML)
	if err != nil {
		return nil, fmt.Errorf("error passing sealing keys as input to kubeseal: %v", err)
	}
	kubesealCommandStdin.Close()
	secretJSON, err := kubesealCommand.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error invoking kubeseal as %v: %v: %s", kubesealCommand.Args, err, secretJSON)
	}

	// build struct
	var secret corev1.Secret
	if err := json.Unmarshal(secretJSON, &secret); err != nil {
		return nil, fmt.Errorf("error unmarshalling json to kubernetes Secret: %v", err)
	}

	// convert .data to .StringData with base64 decoding values
	secret.StringData = map[string]string{}
	for k, v := range secret.Data {
		secret.StringData[k] = string(v)
	}
	// we don't need this anymore
	secret.Data = nil

	// delete metadata.ownerReference
	secret.ObjectMeta.OwnerReferences = nil

	// generate YAML from struct
	secretYAML, err = yaml.Marshal(secret)
	if err != nil {
		return nil, fmt.Errorf("error marshalling kubernetes Secret to YAML: %v", err)
	}

	return secretYAML, nil
}

func Seal(secretYAML []byte, allowEmptyData bool) (sealedSecretYAML []byte, err error) {
	kubesealCommandArgs := []string{
		"-o", "yaml",
	}
	if allowEmptyData {
		kubesealCommandArgs = append(kubesealCommandArgs, "--allow-empty-data")
	}
	kubesealCommand := exec.Command("kubeseal", kubesealCommandArgs...)
	kubesealCommandStdin, _ := kubesealCommand.StdinPipe()
	_, err = kubesealCommandStdin.Write(secretYAML)
	if err != nil {
		return nil, fmt.Errorf("error passing sealing keys as input to kubeseal: %v", err)
	}
	kubesealCommandStdin.Close()
	sealedSecretYAML, err = kubesealCommand.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error invoking kubeseal as %v: %v: %s", kubesealCommand.Args, err, sealedSecretYAML)
	}
	return sealedSecretYAML, nil
}

func EncryptRaw(value []byte, secret corev1.Secret) (encryptedValue []byte, err error) {
	kubesealCommandArgs := []string{
		"--raw",
		"--from-file", "/dev/stdin",
	}
	scope := ssv1alpha1.SecretScope(&secret)
	switch scope {
	case ssv1alpha1.StrictScope:
		kubesealCommandArgs = append(kubesealCommandArgs, "--scope", "strict")
		kubesealCommandArgs = append(kubesealCommandArgs, "--name", secret.Name)
		kubesealCommandArgs = append(kubesealCommandArgs, "--namespace", secret.Namespace)
	case ssv1alpha1.NamespaceWideScope:
		kubesealCommandArgs = append(kubesealCommandArgs, "--scope", "namespace-wide")
		kubesealCommandArgs = append(kubesealCommandArgs, "--namespace", secret.Namespace)
	case ssv1alpha1.ClusterWideScope:
		kubesealCommandArgs = append(kubesealCommandArgs, "--scope", "cluster-wide")
	}
	kubesealCommand := exec.Command("kubeseal", kubesealCommandArgs...)
	kubesealCommandStdin, _ := kubesealCommand.StdinPipe()
	_, err = kubesealCommandStdin.Write(value)
	if err != nil {
		return nil, fmt.Errorf("error passing sealing keys as input to kubeseal: %v", err)
	}
	kubesealCommandStdin.Close()
	encryptedValue, err = kubesealCommand.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error invoking kubeseal as %v: %v: %s", kubesealCommand.Args, err, encryptedValue)
	}
	return encryptedValue, nil
}
