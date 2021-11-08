//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

package extcmd

import (
	"fmt"
	"os"

	"k8s.io/utils/exec"
)

func GetEditor() Editor {
	kc := editor{exec: exec.New()}
	return &kc
}

type Editor interface {
	Edit([]byte) ([]byte, error)
}

type editor struct {
	exec exec.Interface
}

func (e *editor) Edit(content []byte) ([]byte, error) {
	// set editor
	editor := getEnv("VISUAL", getEnv("EDITOR", "vi"))

	// create temporary file for editing
	// TODO: parmeterize file name prefix
	f, err := os.CreateTemp("", "kubectl-sealer-")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file: %v", err)
	}
	defer os.Remove(f.Name())
	f.Write(content)
	f.Close()

	// run editor
	execCmd := e.exec.Command(editor, f.Name())
	execCmd.SetStdin(os.Stdin)
	execCmd.SetStdout(os.Stdout)
	err = execCmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error invoking editor: %v", err)
	}

	editedContent, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}
	return editedContent, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
