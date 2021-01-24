package event

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"testing"
)

func TestEmptyConfigMap(t *testing.T) {
	sns := NewNonFailingSNSClient()
	test.Assertf(t, NewSNSEventNotifier(sns, map[string]string{}) == nil, "expecting nil notifier")
}

func TestNotify(t *testing.T) {
	kv := map[string]string{"topicArn": "some-topic-arn", "region": "some-region"}
	sns := NewNonFailingSNSClient()
	notifier := NewSNSEventNotifier(sns, kv)
	test.Assertf(t, notifier != nil, "expecting valid notifier")
	err := notifier.notify(NewHeartbeatEvent())
	test.Assertf(t, err == nil, "unexpected error : %v", err)
}

func TestBadSNSResponse(t *testing.T) {
	sns := NewFailingSNSClient()
	notifier := NewSNSEventNotifier(sns, map[string]string{"topicArn": "some-topic-arn"})
	test.Assert(t, notifier != nil, "expecting valid notifier")
	err := notifier.notify(NewHeartbeatEvent())
	test.Assertf(t, err != nil, "expecting error response")
}
