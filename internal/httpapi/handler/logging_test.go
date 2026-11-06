package handler

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/rinatkh/classroom14/internal/logging"
)

type failingLogWriter struct {
	header http.Header
	status int
}

func (w *failingLogWriter) Header() http.Header    { return w.header }
func (w *failingLogWriter) WriteHeader(status int) { w.status = status }
func (w *failingLogWriter) Write([]byte) (int, error) {
	return 0, errors.New("private-token-in-writer-error")
}

func TestResponseFailureLogsWithoutLeakingData(t *testing.T) {
	var output bytes.Buffer
	logger, err := logging.New(&output, "http-test", "DEBUG", "json")
	if err != nil {
		t.Fatal(err)
	}
	previous := slog.Default()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(previous) })
	writer := &failingLogWriter{header: make(http.Header)}
	writeJSON(writer, http.StatusOK, map[string]string{"title": "private-body-value"})
	log := output.String()
	if !strings.Contains(log, `"level":"ERROR"`) || strings.Contains(log, "private-") {
		t.Fatalf("unsafe/missing log: %s", log)
	}
	if writer.status != http.StatusOK {
		t.Fatalf("response already started: status=%d", writer.status)
	}
}
