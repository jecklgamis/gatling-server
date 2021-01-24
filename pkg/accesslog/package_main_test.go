package accesslog

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"testing"
)

func TestMain(m *testing.M) {
	test.RunTestAndAssertCoverage(m)
}
