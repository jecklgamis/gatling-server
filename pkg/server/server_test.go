package server

import (
	"fmt"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/waiter"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestServerEndPoints(t *testing.T) {
	os.Setenv("APP_ENVIRONMENT", "dev")
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
