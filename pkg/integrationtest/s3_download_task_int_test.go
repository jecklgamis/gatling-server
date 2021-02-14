package integrationtest

import (
	"encoding/json"
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/env"
	"github.com/jecklgamis/gatling-server/pkg/s3"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/waiter"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDownloadSingleFileSimulationFromS3(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	baseUrl := startServer()
	waiter.WaitUntilHTTPGetOk(fmt.Sprintf("%s/probe/ready", baseUrl), 1*time.Second, 3)
	s3url := env.GetOrPanic("GATLING_SERVER_INCOMING_S3_URL")
	bucket, _, err := s3.ParseS3Uri(s3url)
	test.Assert(t, err == nil, "unable to parse s3 url")

	request := &api.S3DownloadTaskRequest{
		Url:        fmt.Sprintf("s3://%s/SingleFileExampleSimulation.scala", bucket),
		Simulation: "gatling.test.example.simulation.SingleFileExampleSimulation",
		JavaOpts:   "-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPerSecond=1"}

	requestBytes, err := json.Marshal(request)
	test.Assert(t, err == nil, "unable to serialize request")

	url := fmt.Sprintf("%s/task/download/s3", baseUrl)
	reader := strings.NewReader(string(requestBytes))
	resp, err := http.Post(url, "application/json", reader)

	test.Assertf(t, err == nil, "unable to send request to %s", url)
	test.Assertf(t, resp.StatusCode == http.StatusOK, "unable to send request :%v", resp.StatusCode)
	test.Assert(t, resp.Header.Get("Content-Type") == "application/json", "unexpected content type")
	var entity = &api.SubmitTaskResponse{}
	err = json.NewDecoder(resp.Body).Decode(&entity)
	validateArtifacts(t, baseUrl, entity.TaskId)
}
