package sealer

import (
	"bufio"
	"fmt"
	"log"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
)

func EditSecretUntilOK(secretYAML []byte) ([]byte, error) {
	for {
		editedSecretYAML, err := EditWithEditor(secretYAML)
		if err != nil {
			return nil, err
		}

		validationErrors, err := ValidateSecretYAML(editedSecretYAML)
		if err != nil {
			return nil, err
		}

		if len(validationErrors) > 0 {
			log.Printf("validation failed: %v\nPress any key to return to the editor, or Ctrl+C to exit.", validationErrors)
			bufio.NewReader(os.Stdin).ReadByte()
			secretYAML = editedSecretYAML
			continue
		} else {
			return editedSecretYAML, nil
		}
	}
}

func ValidateSecretYAML(secretYAML []byte) (field.ErrorList, error) {
	var secret corev1.Secret
	err := yaml.UnmarshalStrict(secretYAML, &secret)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml to kubernetes Secret: %v", err)
	}

	secret.Data = map[string][]byte{}
	for k, v := range secret.StringData {
		secret.Data[k] = []byte(v)
	}

	return ValidateSecret(&secret), nil
}
