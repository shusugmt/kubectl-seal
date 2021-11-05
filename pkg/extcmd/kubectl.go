//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

package extcmd

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/exec"
)

func GetKubectl() Kubectl {
	kc := kubectl{exec: exec.New()}
	return &kc
}

type Kubectl interface {
	GetSecrets(string, []string) (*corev1.SecretList, error)
}

type kubectl struct {
	exec exec.Interface
}

func (kc *kubectl) GetSecrets(namespace string, labels []string) (*corev1.SecretList, error) {
	// build args
	args := []string{
		"get", "secrets",
		"--output", "json",
	}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	for _, label := range labels {
		args = append(args, "--selector", label)
	}

	// run
	secretJSON, err := kc.exec.Command("kubectl", args...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	// build struct from json
	var secretList corev1.SecretList
	if err := json.Unmarshal(secretJSON, &secretList.ListMeta); err != nil {
		return nil, fmt.Errorf("error unmarshalling json to kubernetes SecretList: %v", err)
	}

	return &secretList, nil
}
