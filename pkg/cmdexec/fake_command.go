package cmdexec

import "os/exec"

type FakeCommandExecutor struct {
	errorToReturn error
}

func NewFakeCommandExecutor(errorToReturn error) *FakeCommandExecutor {
	return &FakeCommandExecutor{errorToReturn}
}

func (f *FakeCommandExecutor) Execute(_ *exec.Cmd) error {
	return f.errorToReturn
}

func (f *FakeCommandExecutor) ExecuteAndLog(_ *exec.Cmd, _ string) error {
	return f.errorToReturn
}
