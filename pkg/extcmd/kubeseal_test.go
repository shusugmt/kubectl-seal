package extcmd

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"testing"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/golang/mock/gomock"
	"github.com/shusugmt/kubectl-sealer/pkg/extcmd/mock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fexec "k8s.io/utils/exec/testing"
)

func TestSealSuccess(t *testing.T) {
	_ = assert.New(t)
	c := gomock.NewController(t)
	defer c.Finish()

	fakeCmd := &fexec.FakeCmd{
		CombinedOutputScript: []fexec.FakeAction{
			func() ([]byte, []byte, error) {
				return []byte(`{"kind":"SealedSecret","apiVersion":"v1alpha1","metadata":{"name":"orange"}}`), []byte(""), nil
			},
		},
	}
	args := []string{"--format", "json"}

	mockExec := mock.NewMockInterface(c)
	mockExec.
		EXPECT().
		Command("kubeseal", args).
		Return(fakeCmd)

	ks := kubeseal{exec: mockExec}
	secret := &corev1.Secret{}
	output, err := ks.Seal(secret)
	if err != nil {
		t.Errorf("test run failed: %v", err)
	}
	expectedOutput := &ssv1alpha1.SealedSecret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1alpha1",
			Kind:       "SealedSecret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "orange",
		},
	}
	assert.Equal(t, expectedOutput, output)

	stdin, err := ioutil.ReadAll(fakeCmd.Stdin)
	if err != nil {
		t.Errorf("test run failed: %v", err)
	}
	expectedStdin, _ := json.Marshal(secret)
	assert.Equal(t, expectedStdin, stdin)
}

func TestUnsealSuccess(t *testing.T) {
	_ = assert.New(t)
	c := gomock.NewController(t)
	defer c.Finish()

	fakeCmd := &fexec.FakeCmd{
		CombinedOutputScript: []fexec.FakeAction{
			func() ([]byte, []byte, error) {
				return []byte(`{"kind":"Secret","apiVersion":"v1","metadata":{"name":"apple"}}`), []byte(""), nil
			},
		},
	}
	args := []interface{}{"--recovery-unseal", "--recovery-private-key", gomock.Any()}

	mockExec := mock.NewMockInterface(c)
	mockExec.
		EXPECT().
		Command("kubeseal", args...).
		Return(fakeCmd)

	ks := kubeseal{exec: mockExec}
	sealedSecret := &ssv1alpha1.SealedSecret{}
	sealingKeys := &corev1.SecretList{}
	output, err := ks.Unseal(sealedSecret, sealingKeys)
	if err != nil {
		t.Errorf("test run failed: %v", err)
	}
	expectedOutput := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apple",
		},
		StringData: map[string]string{},
	}
	assert.Equal(t, expectedOutput, output)

	stdin, err := ioutil.ReadAll(fakeCmd.Stdin)
	if err != nil {
		t.Errorf("test run failed: %v", err)
	}
	expectedStdin, _ := json.Marshal(sealedSecret)
	assert.Equal(t, expectedStdin, stdin)
}

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

func TestEncryptRawFail(t *testing.T) {
	_ = assert.New(t)
	for _, td := range []struct {
		subject      string
		secret       *corev1.Secret
		expectedArgs []string
		err          string
	}{
		{
			subject: "kubeseal returned error",
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
	} {
		t.Run(td.subject, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			fakeCmd := &fexec.FakeCmd{
				CombinedOutputScript: []fexec.FakeAction{
					func() ([]byte, []byte, error) {
						err := &exec.ExitError{}
						return []byte(""), []byte(`error: cannot fetch certificate: services "sealed-secrets-controller" not found`), err
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
			_, err := ks.EncryptRaw([]byte(""), td.secret)
			assert.Error(t, err)
		})
	}
}

func TestEncryptRawFailArgs(t *testing.T) {
	_ = assert.New(t)
	for _, td := range []struct {
		subject string
		secret  *corev1.Secret
		err     string
	}{
		{
			subject: "StrictScope, namespace not given",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "apple",
					Annotations: ssv1alpha1.UpdateScopeAnnotations(
						map[string]string{},
						ssv1alpha1.StrictScope,
					),
				},
			},
			err: "namespace must be given",
		},
		{
			subject: "StrictScope, name not given",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "banana",
					Annotations: ssv1alpha1.UpdateScopeAnnotations(
						map[string]string{},
						ssv1alpha1.StrictScope,
					),
				},
			},
			err: "name must be given",
		},
		{
			subject: "NamespaceWideScope, namespace not given",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "apple",
					Annotations: ssv1alpha1.UpdateScopeAnnotations(
						map[string]string{},
						ssv1alpha1.NamespaceWideScope,
					),
				},
			},
			err: "namespace must be given",
		},
	} {
		t.Run(td.subject, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			mockExec := mock.NewMockInterface(c)
			ks := kubeseal{exec: mockExec}
			_, err := ks.EncryptRaw([]byte(""), td.secret)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), td.err)
			}
		})
	}
}
