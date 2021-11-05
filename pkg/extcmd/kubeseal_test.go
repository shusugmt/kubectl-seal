package extcmd

import (
	"io/ioutil"
	"testing"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/golang/mock/gomock"
	"github.com/shusugmt/kubectl-sealer/pkg/extcmd/mock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fexec "k8s.io/utils/exec/testing"
)

func TestEncryptRawSuccess(t *testing.T) {
	_ = assert.New(t)
	for _, td := range []struct {
		subject      string
		value        []byte
		secret       *corev1.Secret
		expectedArgs []string
	}{
		{
			subject: "StrictScope",
			value:   []byte("orange"),
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "apple",
					Namespace: "banana",
					Annotations: ssv1alpha1.UpdateScopeAnnotations(
						map[string]string{},
						ssv1alpha1.StrictScope,
					),
				},
			},
			expectedArgs: []string{
				"--scope", "strict",
				"--name", "apple",
				"--namespace", "banana",
			},
		},
		{
			subject: "NamespaceWideScope",
			value:   []byte("orange"),
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "apple",
					Namespace: "banana",
					Annotations: ssv1alpha1.UpdateScopeAnnotations(
						map[string]string{},
						ssv1alpha1.NamespaceWideScope,
					),
				},
			},
			expectedArgs: []string{
				"--scope", "namespace-wide",
				"--namespace", "banana",
			},
		},
		{
			subject: "ClusterWideScope",
			value:   []byte("orange"),
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "apple",
					Namespace: "banana",
					Annotations: ssv1alpha1.UpdateScopeAnnotations(
						map[string]string{},
						ssv1alpha1.ClusterWideScope,
					),
				},
			},
			expectedArgs: []string{
				"--scope", "cluster-wide",
			},
		},
	} {
		t.Run(td.subject, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			fakeCmd := &fexec.FakeCmd{
				CombinedOutputScript: []fexec.FakeAction{
					func() ([]byte, []byte, error) {
						return []byte("ORANGE"), []byte(""), nil
					},
				},
			}
			args := append([]string{"--raw", "--from-file", "/dev/stdin"}, td.expectedArgs...)

			mockExec := mock.NewMockInterface(c)
			mockExec.
				EXPECT().
				Command("kubeseal", args).
				Return(fakeCmd)

			ks := kubeseal{exec: mockExec}
			output, err := ks.EncryptRaw(td.value, td.secret)
			if err != nil {
				t.Errorf("test run failed: %v", err)
			}
			assert.Equal(t, []byte("ORANGE"), output)

			stdin, err := ioutil.ReadAll(fakeCmd.Stdin)
			if err != nil {
				t.Errorf("test run failed: %v", err)
			}
			assert.Equal(t, td.value, stdin)
		})
	}
}
