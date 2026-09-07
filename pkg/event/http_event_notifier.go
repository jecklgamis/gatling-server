package event

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/jsonutil"
	"log/slog"
	"net/http"
	"strings"
)

type HttpEventNotifier struct {
	ConfigMap map[string]string
}

func NewHTTPNotifier(configMap map[string]string) *HttpEventNotifier {
	if _, ok := configMap["url"]; !ok {
		slog.Warn("No url found in config map")
		return nil
	}
	return &HttpEventNotifier{ConfigMap: configMap}
}

func (h *HttpEventNotifier) Event(event interface{}) {
	if err := h.notify(event); err != nil {
		slog.Error("Unable to notify event", "error", err)
	}
}

func (h *HttpEventNotifier) notify(event interface{}) error {
	resp, err := http.Post(h.ConfigMap["url"], "application/json", strings.NewReader(jsonutil.ToJson(event)))
	if err != nil {
		slog.Error("Failed sending HTTP request", "error", err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Unable to close response body", "error", err)
		}
	}()
	slog.Info("Sent HTTP request", "event", jsonutil.ToJson(event))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}
