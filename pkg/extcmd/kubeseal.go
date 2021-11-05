//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

package extcmd

import (
	"bytes"
	"fmt"

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
	Seal(secret *corev1.Secret) (*ssv1alpha1.SealedSecret, error)
	Unseal(secret *corev1.SecretList) (*corev1.Secret, error)
	EncryptRaw([]byte, *corev1.Secret) ([]byte, error)
}

type kubeseal struct {
	exec exec.Interface
}

func (ks *kubeseal) Seal(secret *corev1.Secret) (*ssv1alpha1.SealedSecret, error) {
	return nil, nil
}

func (ks *kubeseal) Unseal(sealingKeys *corev1.SecretList) (*corev1.Secret, error) {
	return nil, nil
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
