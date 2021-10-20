package sealer

import (
	"fmt"
	"os"
	"os/exec"
)

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// take content, open editor, then return edited content
// a.k.a. vipe
func EditWithEditor(content []byte) (editedContent []byte, err error) {

	// set editor
	editor := GetEnv("VISUAL", GetEnv("EDITOR", "vi"))

	// create and open a temporary file for editing
	f, err := os.CreateTemp("", "kubectl-sealer-")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file: %v", err)
	}
	// remove the temporary file at the end of the program
	defer os.Remove(f.Name())

	// write content and close immediately so that we can run editor
	f.Write(content)
	f.Close()

	// run editor
	editorCommand := exec.Command(editor, f.Name())
	editorCommand.Stdin = os.Stdin
	editorCommand.Stdout = os.Stdout
	err = editorCommand.Run()
	if err != nil {
		return nil, fmt.Errorf("error invoking editor as %v: %v", editorCommand.Args, err)
	}

	// read content after editing
	editedContent, err = os.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}

	return editedContent, nil
}

// given map A and B, returns list of keys only exists in map A
// if there is no such key, returns empty slice
func GetKeyDiff(a map[string]string, b map[string]string) (keys []string) {
	keys = []string{}
	for aKey := range a {
		// if no matching key found in B = this key only exists in A
		if _, ok := b[aKey]; !ok {
			keys = append(keys, aKey)
		}
	}
	return keys
}

// given map A and B, returns a map such that
// - key exists in both A and B
// - value of A(key) differs B(key)
func GetUpdatedExisting(a map[string]string, b map[string]string) (m map[string]string) {
	m = map[string]string{}
	for aKey, aVal := range a {
		// if matching key found in B = this key exists in both A and B
		if bVal, ok := b[aKey]; ok {
			if aVal != bVal {
				m[aKey] = aVal
			}
		}
	}
	return m
}
