package integrationtest

import (
	"encoding/json"
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/env"
	"github.com/jecklgamis/gatling-server/pkg/server"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/uploader"
	"github.com/jecklgamis/gatling-server/pkg/waiter"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func startServer() (baseUrl string) {
	os.Setenv("APP_ENVIRONMENT", "dev")
	port := test.UnusedPort()
	uploadDir, _ := ioutil.TempDir("", "uploads")
	workspaceDir, _ := ioutil.TempDir("", "repos")
	go func() {
		viper.Set("SERVER.HTTP.PORT", fmt.Sprintf("%d", port))
		viper.Set("SERVER.HTTPS.PORT", fmt.Sprintf("%d", test.UnusedPort()))
		viper.Set("SCRIPTSDIR", env.GetOrElse("GATLING_TEST_SCRIPTS_DIR", "../../scripts"))
		viper.Set("UPLOADDIR", uploadDir)
		viper.Set("WORKSPACEDIR", workspaceDir)
		server.Start()
	}()
	baseUrl = fmt.Sprintf("http://localhost:%d", port)
	waiter.WaitUntilHTTPGetOk(fmt.Sprintf("%s/probe/ready", baseUrl), 1*time.Second, 3)
	return baseUrl
}

func TestSubmitJarSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	baseUrl := startServer()
	waiter.WaitUntilHTTPGetOk(fmt.Sprintf("%s/probe/ready", baseUrl), 1*time.Second, 3)

	kv := map[string]string{
		"simulation": "gatling.test.example.simulation.ExampleSimulation",
		"javaOpts":   "-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPersecond=1",
	}
	uploadUrl := fmt.Sprintf("%s/task/upload/http", baseUrl)
	resp, err := uploader.UploadFile(uploadUrl, "testdata/gatling-scala-example-lean.jar", kv)
	test.Assertf(t, err == nil, "unable to upload : %v", err)
	test.Assert(t, resp.StatusCode == http.StatusOK, "expecting 200")
	var entity = &api.SubmitTaskResponse{}
	err = json.NewDecoder(resp.Body).Decode(&entity)
	test.Assertf(t, err == nil, "unable to parse response : %v", err)
	validateArtifacts(t, baseUrl, entity.TaskId)

}

func TestAbortTask(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	baseUrl := startServer()
	waiter.WaitUntilHTTPGetOk(fmt.Sprintf("%s/probe/ready", baseUrl), 1*time.Second, 3)

	kv := map[string]string{
		"simulation": "gatling.test.example.simulation.ExampleSimulation",
		"javaOpts":   "-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPersecond=1",
	}
	uploadUrl := fmt.Sprintf("%s/task/upload/http", baseUrl)
	resp, err := uploader.UploadFile(uploadUrl, "testdata/gatling-scala-example-lean.jar", kv)
	test.Assertf(t, err == nil, "unable to upload : %v", err)
	test.Assert(t, resp.StatusCode == http.StatusOK, "expecting 200")
	var entity = &api.SubmitTaskResponse{}
	err = json.NewDecoder(resp.Body).Decode(&entity)
	test.Assertf(t, err == nil, "unable to parse response : %v", err)

	time.Sleep(3 * time.Second)
	abortUrl := fmt.Sprintf("%s/task/abort/%s", baseUrl, entity.TaskId)
	log.Println(abortUrl)
	resp, err = http.Post(abortUrl, "*/*", strings.NewReader(""))
	test.Assertf(t, err == nil, "unable to abort task :%v", err)
	test.Assertf(t, resp.StatusCode == http.StatusOK, "expecting 200 but got %v", resp.StatusCode)
}

func validateArtifacts(t *testing.T, baseUrl string, taskId string) {
	err := waiter.WaitUntilHTTPGetOk(taskUrl(baseUrl, "metadata", taskId), 1*time.Second, 45)
	test.Assertf(t, err == nil, "failed to fetch metadata")

	err = waiter.WaitUntilHTTPGetOk(taskUrl(baseUrl, "console", taskId), 1*time.Second, 45)
	test.Assertf(t, err == nil, "failed to fetch console logs")

	err = waiter.WaitUntilHTTPGetOk(taskUrl(baseUrl, "results", taskId), 1*time.Second, 45)
	test.Assertf(t, err == nil, "failed to fetch results")

	err = waiter.WaitUntilHTTPGetOk(taskUrl(baseUrl, "simulationLog", taskId), 1*time.Second, 45)
	test.Assertf(t, err == nil, "failed to fetch simulation log")

	err = waiter.WaitUntilHTTPGetOk(taskUrl(baseUrl, "", taskId), 1*time.Second, 45)
	test.Assertf(t, err == nil, "failed to fetch task")
}

func taskUrl(baseUrl, path string, taskId string) string {
	return fmt.Sprintf("%s/task/%s/%s", baseUrl, path, taskId)
}
