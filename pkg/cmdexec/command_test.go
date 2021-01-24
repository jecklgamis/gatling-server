package cmdexec

import (
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"io/ioutil"
	"os/exec"
	"testing"
)

func TestExecuteInvalidCommand(t *testing.T) {
	err := NewCommandExecutor().Execute(exec.Command("blah"))
	test.Assertf(t, err != nil, "expecting invalid command error")
}

func TestExecuteCommand(t *testing.T) {
	err := NewCommandExecutor().Execute(exec.Command("ls", "-la"))
	test.Assertf(t, err == nil, "unexpected error :%v", err)
}

func TestExecuteAndLog(t *testing.T) {
	console, err := ioutil.TempFile("", "console.log")
	test.Assertf(t, err == nil, "unable to create file :%v", err)
	err = NewCommandExecutor().ExecuteAndLog(exec.Command("ifconfig"), console.Name())
	test.Assertf(t, err == nil, "unable to execute command :%v", err)
	test.Assertf(t, fileioutil.FileExist(console.Name()), "failed to execute command")
}
