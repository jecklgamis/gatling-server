package testing

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
)

func Assert(t *testing.T, condition bool, message string) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: %s", filepath.Base(file), line, message)
	}
}

func Assertf(t *testing.T, condition bool, format string, args ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: %s", filepath.Base(file), line, fmt.Sprintf(format, args...))
	}
}
