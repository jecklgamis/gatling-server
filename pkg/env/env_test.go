package env

import (
	"fmt"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"os"
	"testing"
)

func TestGetOrElse(t *testing.T) {
	os.Setenv("some-key", "some-value")
	if v := GetOrElse("some-key", ""); v != "some-value" {
		t.Errorf("Got %s, expecting %s", v, "some-default")
	}
}

func TestGetOrElseFallback(t *testing.T) {
	if v := GetOrElse("some-other-key", "some-default"); v != "some-default" {
		t.Errorf("Got %s, expecting %s", v, "some-default")
	}
}

func TestGetOrPanic(t *testing.T) {
	var panicCaught = false
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			panicCaught = true
		}
		test.Assertf(t, panicCaught, "expecting panic")
	}()
	GetOrPanic("some-non-existing-key")
}
