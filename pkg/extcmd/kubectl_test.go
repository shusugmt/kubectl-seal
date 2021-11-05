package extcmd

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shusugmt/kubectl-sealer/pkg/extcmd/mock"
	"github.com/stretchr/testify/assert"
	fexec "k8s.io/utils/exec/testing"
)

func TestGetSecretsSuccess(t *testing.T) {
	_ = assert.New(t)
	for _, td := range []struct {
		subject      string
		namespace    string
		labels       []string
		expectedArgs []string
	}{
		{
			subject:      "no namespace, no labels",
			namespace:    "",
			labels:       nil,
			expectedArgs: []string{"get", "secrets", "--output", "json"},
		},
		{
			subject:      "with namespace, no labels",
			namespace:    "default",
			labels:       nil,
			expectedArgs: []string{"get", "secrets", "--output", "json", "--namespace", "default"},
		},
		{
			subject:   "no namespace, with 2 labels",
			namespace: "",
			labels:    []string{"app.kubernetes.io/name", "app.kubernetes.io/component=db"},
			expectedArgs: []string{
				"get", "secrets", "--output", "json",
				"--selector", "app.kubernetes.io/name", "--selector", "app.kubernetes.io/component=db",
			},
		},
		{
			subject:   "with namespace, with 2 labels",
			namespace: "default",
			labels:    []string{"app.kubernetes.io/name", "app.kubernetes.io/component=db"},
			expectedArgs: []string{
				"get", "secrets", "--output", "json",
				"--namespace", "default",
				"--selector", "app.kubernetes.io/name", "--selector", "app.kubernetes.io/component=db",
			},
		},
	} {
		t.Run(td.subject, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			mockExec := mock.NewMockInterface(c)
			mockExec.
				EXPECT().
				Command("kubectl", td.expectedArgs).
				Return(&fexec.FakeCmd{
					CombinedOutputScript: []fexec.FakeAction{
						func() ([]byte, []byte, error) {
							return []byte(`{"apiVersion":"v1","items":[],"kind":"List","metadata":{"resourceVersion":"","selfLink":""}}`), []byte(""), nil
						},
					},
				})

			kc := kubectl{exec: mockExec}
			_, err := kc.GetSecrets(td.namespace, td.labels)
			if err != nil {
				t.Errorf("test run failed: %v", err)
			}
		})
	}
}
