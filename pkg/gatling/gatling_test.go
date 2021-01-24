package gatling

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/cmdexec"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func internalServerErrorHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestRunSimulationWithSuccessfulResult(t *testing.T) {
	task := someGatlingTask(t, "http://localhost:8080")
	_, err := SomeGatlingDist().RunSimulation(cmdexec.NewFakeCommandExecutor(nil), task)
	test.Assert(t, err == nil, "expecting nil return")
}

func TestRunSimulationWithFailedResult(t *testing.T) {
	task := someGatlingTask(t, "http://some-target-url")
	_, err := SomeGatlingDist().RunSimulation(cmdexec.NewFakeCommandExecutor(fmt.Errorf("some-error")), task)
	test.Assert(t, err != nil, "expecting error return")
}

func validateUserFilesDirArtifacts(t *testing.T, userFilesDir *workspace.UserFilesDir) {
	test.Assertf(t, fileioutil.FileExist(filepath.Join(userFilesDir.BaseDir, "console.log")),
		"expecting console.log")
	_, err := fileioutil.FindFile(userFilesDir.BaseDir, "simulation.log")
	test.Assertf(t, err == nil, "simulation.log not found")
}

func TestRunSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	handlers := map[bool]http.Handler{
		true:  http.HandlerFunc(okHandler),
		false: http.HandlerFunc(internalServerErrorHandler)}

	gatling := SomeGatlingDist()
	for expectedOk, handler := range handlers {
		server := httptest.NewServer(handler)
		task := someGatlingTask(t, server.URL)
		cmd, err := gatling.RunSimulation(cmdexec.NewCommandExecutor(), task)
		test.Assertf(t, err == nil, "failed to start simulation")
		err = cmd.Wait()
		if expectedOk {
			test.Assertf(t, err == nil, "expecting simulation to succeed")
		} else {
			test.Assertf(t, err != nil, "expecting simulation to fail")
		}
		if expectedOk {
			validateUserFilesDirArtifacts(t, task.UserFilesDir)
		}
	}
}

func TestAbortSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	server := httptest.NewServer(http.HandlerFunc(okHandler))
	task := someGatlingTask(t, fmt.Sprintf(server.URL))
	cmd, _ := SomeGatlingDist().RunSimulation(cmdexec.NewCommandExecutor(), task)
	go func() {
		time.Sleep(3 * time.Second)
		cmd.Process.Kill()
	}()
	err := cmd.Wait()
	test.Assertf(t, strings.Contains(err.Error(), "signal: killed"), "expecting killed process")
}

func someGatlingTask(t *testing.T, targetUrl string) *Task {
	userFilesPath, _ := ioutil.TempDir("", "")
	userFilesDir, _ := workspace.NewUserFilesDir(filepath.Join(userFilesPath, "user-files-dir"))
	test.Assertf(t, userFilesDir != nil, "nil user files dir")
	fileioutil.CopyFile("testdata/SingleFileExampleSimulation.scala",
		fmt.Sprintf("%s/SingleFileExampleSimulation.scala",
			userFilesDir.Simulations))
	return NewTask("some-task-id", "gatling.test.example.simulation.SingleFileExampleSimulation",
		fmt.Sprintf("-DbaseUrl=%s -DdurationMin=0.10 -DrequestPerSecond=1", targetUrl), userFilesDir)
}
