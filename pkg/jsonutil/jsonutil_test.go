package jsonutil

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"testing"
)

func TestSerialize(t *testing.T) {
	kv := map[string]string{"someKey": "someValue"}
	j := ToJson(kv)
	bytes := []byte(`{"someKey":"someValue"}`)
	test.Assertf(t, j == string(bytes), "expecting {} but got %v", j)
}

func TestSerializeNil(t *testing.T) {
	test.Assertf(t, ToJson(nil) == "{}", "expecting {}")
}
