package testing

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var MinimumCodeCoverage = 0.70

func RunTestAndAssertCoverage(m *testing.M) {
	flag.Parse()
	rc := m.Run()
	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < MinimumCodeCoverage {
			fmt.Printf("Test coverage failed, expecting %v%% but got %0.2f%%\n",
				MinimumCodeCoverage*100, c*100)
			rc = -1
		}
	}
	os.Exit(rc)
}
