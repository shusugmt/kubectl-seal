//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

package extcmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/exec"
)

func GetKubeseal() Kubeseal {
	ks := kubeseal{
		exec: exec.New(),
	}
	return &ks
}

type Kubeseal interface {
	Seal(*corev1.Secret) (*ssv1alpha1.SealedSecret, error)
	Unseal(*ssv1alpha1.SealedSecret, *corev1.SecretList) (*corev1.Secret, error)
	EncryptRaw([]byte, *corev1.Secret) ([]byte, error)
}

type kubeseal struct {
	exec exec.Interface
}

func (ks *kubeseal) Seal(secret *corev1.Secret) (*ssv1alpha1.SealedSecret, error) {
	args := []string{
		"--format", "json",
	}
	execCmd := ks.exec.Command("kubeseal", args...)

	secretJSON, err := json.Marshal(secret)
	if err != nil {
		return nil, fmt.Errorf("error marshalling kubernetes Secret to json: %v", err)
	}
	execCmd.SetStdin(bytes.NewReader(secretJSON))

	sealedSecretJSON, err := execCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error invoking kubeseal as %v: %v: %s", args, err, sealedSecretJSON)
	}

	// build struct
	var sealedSecret ssv1alpha1.SealedSecret
	if err := json.Unmarshal(sealedSecretJSON, &sealedSecret); err != nil {
		return nil, fmt.Errorf("error unmarshalling json to SealedSecret: %v", err)
	}
	return &sealedSecret, nil
}

func (ks *kubeseal) Unseal(sealedSecret *ssv1alpha1.SealedSecret, sealingKeys *corev1.SecretList) (*corev1.Secret, error) {
	// create temporary file to store sealing keys
	// TODO: parmeterize file name prefix
	f, err := os.CreateTemp("", "kubectl-sealer-")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file: %v", err)
	}
	defer os.Remove(f.Name())

	sealingKeysJSON, err := json.Marshal(sealingKeys)
	if err != nil {
		return nil, fmt.Errorf("error marshalling kubernetes SecretList to json: %v", err)
	}
	_, err = f.Write(sealingKeysJSON)
	if err != nil {
		return nil, fmt.Errorf("error writing kubernetes SecretList to file: %v: %v", f, err)
	}
	err = f.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing file: %v: %v", f, err)
	}

	// do unseal
	args := []string{
		"--recovery-unseal",
		"--recovery-private-key", f.Name(),
	}
	execCmd := ks.exec.Command("kubeseal", args...)

	sealedSecretJSON, err := json.Marshal(sealedSecret)
	if err != nil {
		return nil, fmt.Errorf("error marshalling SealedSecret to json: %v", err)
	}
	execCmd.SetStdin(bytes.NewReader(sealedSecretJSON))

	secretJSON, err := execCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error invoking kubeseal as %v: %v: %s", args, err, secretJSON)
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

	return &secret, nil
}

func (ks *kubeseal) EncryptRaw(value []byte, secret *corev1.Secret) ([]byte, error) {
	args := []string{
		"--raw",
		"--from-file", "/dev/stdin",
	}

	scope := ssv1alpha1.SecretScope(secret)
	switch scope {
	case ssv1alpha1.StrictScope:
		if secret.Name == "" {
			return nil, fmt.Errorf("name must be given")
		}
		if secret.Namespace == "" {
			return nil, fmt.Errorf("namespace must be given")
		}
		args = append(args, "--scope", "strict")
		args = append(args, "--name", secret.Name)
		args = append(args, "--namespace", secret.Namespace)
	case ssv1alpha1.NamespaceWideScope:
		if secret.Namespace == "" {
			return nil, fmt.Errorf("namespace must be given")
		}
		args = append(args, "--scope", "namespace-wide")
		args = append(args, "--namespace", secret.Namespace)
	case ssv1alpha1.ClusterWideScope:
		args = append(args, "--scope", "cluster-wide")
	}

	execCmd := ks.exec.Command("kubeseal", args...)
	execCmd.SetStdin(bytes.NewReader(value))
	encryptedValue, err := execCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error invoking kubeseal as %v: %v: %s", args, err, encryptedValue)
	}
	return encryptedValue, nil
}
