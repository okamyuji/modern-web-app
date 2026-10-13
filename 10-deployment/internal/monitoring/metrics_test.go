package monitoring

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func seriesCount(t *testing.T, family string) int {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == family {
			return len(f.GetMetric())
		}
	}
	return 0
}

func TestPrometheusMiddleware_BoundsLabelCardinality(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/known", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/", http.NotFound)
	h := PrometheusMiddleware(mux)
	before := seriesCount(t, "http_requests_total")

	for i := range 50 {
		req := httptest.NewRequest(fmt.Sprintf("M%d", i), fmt.Sprintf("/random/%d", i), strings.NewReader("x"))
		h.ServeHTTP(httptest.NewRecorder(), req)
	}

	if added := seriesCount(t, "http_requests_total") - before; added > 1 {
		t.Fatalf("50 random paths and methods added %d series, want at most 1", added)
	}
}

func TestPrometheusMiddleware_LabelsByRoutePattern(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {})
	h := PrometheusMiddleware(mux)
	labels := map[string]string{"method": "GET", "endpoint": "/items/", "status_code": "200"}
	before := testCounterValue(t, "http_requests_total", labels)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items/1", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items/2", nil))

	if got := testCounterValue(t, "http_requests_total", labels) - before; got != 2 {
		t.Fatalf("GET /items/ count increased by %v, want 2", got)
	}
}

func testCounterValue(t *testing.T, family string, labels map[string]string) float64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() != family {
			continue
		}
	metrics:
		for _, m := range f.GetMetric() {
			for _, lp := range m.GetLabel() {
				if want, ok := labels[lp.GetName()]; ok && want != lp.GetValue() {
					continue metrics
				}
			}
			return m.GetCounter().GetValue()
		}
	}
	return 0
}

func TestMetricMethod(t *testing.T) {
	cases := map[string]string{
		http.MethodGet:    http.MethodGet,
		http.MethodPost:   http.MethodPost,
		http.MethodDelete: http.MethodDelete,
		"PROPFIND":        "OTHER",
		"m1":              "OTHER",
	}
	for in, want := range cases {
		if got := metricMethod(in); got != want {
			t.Errorf("metricMethod(%q) = %q, want %q", in, got, want)
		}
	}
}

func histogramCount(t *testing.T, family, endpoint string) uint64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() != family {
			continue
		}
		for _, m := range f.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "endpoint" && lp.GetValue() == endpoint {
					return m.GetHistogram().GetSampleCount()
				}
			}
		}
	}
	return 0
}

func TestPrometheusMiddleware_RecordsRequestSizeOnlyWithBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {})
	h := PrometheusMiddleware(mux)
	before := histogramCount(t, "http_request_size_bytes", "/upload")

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("")))
	afterEmpty := histogramCount(t, "http_request_size_bytes", "/upload")
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("x")))
	afterBody := histogramCount(t, "http_request_size_bytes", "/upload")

	if afterEmpty != before {
		t.Fatalf("empty body recorded a size sample (%d -> %d)", before, afterEmpty)
	}
	if afterBody != before+1 {
		t.Fatalf("1-byte body samples = %d, want %d", afterBody, before+1)
	}
}
