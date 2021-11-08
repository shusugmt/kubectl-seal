package extcmd

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shusugmt/kubectl-sealer/pkg/extcmd/mock"
	"github.com/stretchr/testify/assert"
	fexec "k8s.io/utils/exec/testing"
)

func TestEditSuccess(t *testing.T) {
	_ = assert.New(t)
	c := gomock.NewController(t)
	defer c.Finish()

	fakeCmd := &fexec.FakeCmd{
		RunScript: []fexec.FakeAction{
			func() ([]byte, []byte, error) {
				return nil, nil, nil
			},
		},
	}
	args := []interface{}{gomock.Any()}

	mockExec := mock.NewMockInterface(c)
	mockExec.
		EXPECT().
		Command("vi", args...).
		Do(func(argc string, argv ...string) {
			tmpfile := argv[len(argv)-1]
			os.WriteFile(tmpfile, []byte("after"), 0644)
		}).
		Return(fakeCmd)

	editor := editor{exec: mockExec}
	output, err := editor.Edit([]byte("before"))
	if err != nil {
		t.Errorf("test run failed: %v", err)
	}
	assert.Equal(t, []byte("after"), output)
}
