package server

import (
	"fmt"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/waiter"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// ensureTestCerts generates the self-signed TLS cert/key pair config-dev.yaml
// points HTTPS at (testdata/server.{key,crt}), if not already present. These
// are not committed to the repo, so they must be generated locally.
func ensureTestCerts(t *testing.T) {
	if _, err := os.Stat("testdata/server.key"); err == nil {
		return
	}
	cmd := exec.Command("../../scripts/generate-ssl-certs.sh", "testdata")
	test.Assertf(t, cmd.Run() == nil, "unable to generate test TLS certs")
}

func TestServerEndPoints(t *testing.T) {
	ensureTestCerts(t)
	test.Assertf(t, os.Setenv("APP_ENVIRONMENT", "dev") == nil, "unable to set env var")
	test.Assertf(t, os.Setenv("API_TOKEN", "some-test-api-token") == nil, "unable to set env var")
	port := test.UnusedPort()
	go func() {
		viper.Set("SERVER.HTTP.PORT", fmt.Sprintf("%d", port))
		viper.Set("SERVER.HTTPS.PORT", fmt.Sprintf("%d", test.UnusedPort()))
		Start()
	}()
	baseUrl := fmt.Sprintf("http://localhost:%d/", port)
	err := waiter.WaitUntilHTTPGetOk(baseUrl, 1*time.Second, 10)
	test.Assertf(t, err == nil, "server down :%v", err)

	r, err := http.Get(fmt.Sprintf("%s/buildInfo", baseUrl))
	test.Assertf(t, err == nil, "unable to send request : %v", err)
	test.Assert(t, r.StatusCode == http.StatusOK, "unable to reach /buildInfo")
	test.Assert(t, r.Header.Get("Content-Type") == "application/json", "unexpected Content-Type from /buildInfo")

	r, _ = http.Get(fmt.Sprintf("%s/probe/ready", baseUrl))
	test.Assertf(t, err == nil, "unable to send request : %v", err)
	test.Assert(t, r.StatusCode == http.StatusOK, "unable to reach /probe/ready")
	test.Assert(t, r.Header.Get("Content-Type") == "application/json", "unexpected Content-Type from /probe/ready")

	r, _ = http.Get(fmt.Sprintf("%s/probe/live", baseUrl))
	test.Assertf(t, err == nil, "unable to send request : %v", err)
	test.Assert(t, r.StatusCode == http.StatusOK, "unable to reach /probe/live")
	test.Assert(t, r.Header.Get("Content-Type") == "application/json", "unexpected Content-Type from /probe/live")

	r, _ = http.Get(fmt.Sprintf("%s/blackhole", baseUrl))
	test.Assertf(t, err == nil, "unable to send request : %v", err)
	test.Assert(t, r.StatusCode == http.StatusOK, "unable to reach /blackhole")
}

func TestAccessLogWritesToConfiguredFile(t *testing.T) {
	ensureTestCerts(t)
	test.Assertf(t, os.Setenv("APP_ENVIRONMENT", "dev") == nil, "unable to set env var")
	test.Assertf(t, os.Setenv("API_TOKEN", "some-test-api-token") == nil, "unable to set env var")
	port := test.UnusedPort()
	accessLogFile := fmt.Sprintf("%s/access-%d.log", t.TempDir(), port)
	go func() {
		viper.Set("SERVER.HTTP.PORT", fmt.Sprintf("%d", port))
		viper.Set("SERVER.HTTPS.PORT", fmt.Sprintf("%d", test.UnusedPort()))
		viper.Set("ACCESSLOG.ENABLED", "true")
		viper.Set("ACCESSLOG.FILE", accessLogFile)
		Start()
	}()
	baseUrl := fmt.Sprintf("http://localhost:%d/", port)
	err := waiter.WaitUntilHTTPGetOk(baseUrl, 1*time.Second, 10)
	test.Assertf(t, err == nil, "server down :%v", err)

	r, err := http.Get(fmt.Sprintf("%s/buildInfo", baseUrl))
	test.Assertf(t, err == nil, "unable to send request : %v", err)
	test.Assert(t, r.StatusCode == http.StatusOK, "unable to reach /buildInfo")

	var content []byte
	err = waiter.WaitUntil(200*time.Millisecond, 10, func(counter int) bool {
		content, err = os.ReadFile(accessLogFile)
		return err == nil && len(content) > 0
	})
	test.Assertf(t, err == nil, "access log file was not written to :%v", err)
	test.Assertf(t, strings.Contains(string(content), `"msg":"access"`), "expecting access entry, got %q", string(content))
	test.Assertf(t, strings.Contains(string(content), `"uri_path":"/buildInfo"`), "expecting uri_path, got %q", string(content))
}
