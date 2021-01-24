package testing

import (
	"testing"
)

func TestUnusedPort(t *testing.T) {
	Assert(t, UnusedPort() > 1024, "Port must be > 1024")
}
