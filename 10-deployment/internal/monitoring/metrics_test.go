package monitoring

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

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

func TestStartMetricsServer_ReturnsErrorForInvalidAddr(t *testing.T) {
	for _, addr := range []string{"[[::1]]:0", "no-such-host.invalid:0", "127.0.0.1:9999:0"} {
		t.Run(addr, func(t *testing.T) {
			if _, err := StartMetricsServer(addr, time.Minute); err == nil {
				t.Fatalf("StartMetricsServer(%q) = nil error, want listen error", addr)
			}
		})
	}
}

func TestStartMetricsServer_ServesMetrics(t *testing.T) {
	ln, err := StartMetricsServer("127.0.0.1:0", time.Minute)
	if err != nil {
		t.Fatalf("StartMetricsServer: %v", err)
	}

	resp, err := http.Get("http://" + ln.String() + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestStartMetricsServer_ServesHealth(t *testing.T) {
	ln, err := StartMetricsServer("127.0.0.1:0", time.Minute)
	if err != nil {
		t.Fatalf("StartMetricsServer: %v", err)
	}

	resp, err := http.Get("http://" + ln.String() + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK || string(body) != "OK" {
		t.Fatalf("status = %d, body = %q, want 200 OK", resp.StatusCode, body)
	}
}

func TestStartMetricsServer_ClosesIdleKeepAliveConnections(t *testing.T) {
	addr, err := StartMetricsServer("127.0.0.1:0", 200*time.Millisecond)
	if err != nil {
		t.Fatalf("StartMetricsServer: %v", err)
	}
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	fmt.Fprint(conn, "GET /health HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if _, err := http.ReadResponse(bufio.NewReader(conn), nil); err != nil {
		t.Fatalf("first response: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Read(make([]byte, 1))

	if !errors.Is(err, io.EOF) {
		t.Fatalf("read after idle = %v, want EOF from the server closing the idle connection", err)
	}
}

func TestServeUntilSignal_WaitsForInFlightRequest(t *testing.T) {
	started := make(chan struct{})
	var finished atomic.Bool
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(300 * time.Millisecond)
		finished.Store(true)
		w.Write([]byte("done"))
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	quit := make(chan os.Signal, 1)
	var cleaned atomic.Bool
	served := make(chan error, 1)
	go func() {
		served <- ServeUntilSignal(&http.Server{Handler: mux}, ln, quit, 5*time.Second, func() { cleaned.Store(true) })
	}()
	got := make(chan string, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/slow")
		if err != nil {
			got <- err.Error()
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		got <- string(body)
	}()
	<-started

	quit <- syscall.SIGTERM
	err = <-served

	if err != nil {
		t.Fatalf("ServeUntilSignal = %v, want nil", err)
	}
	if !finished.Load() || !cleaned.Load() {
		t.Fatalf("returned before the in-flight request finished (%v) or cleanup ran (%v)", finished.Load(), cleaned.Load())
	}
	if body := <-got; body != "done" {
		t.Fatalf("client got %q, want done", body)
	}
}

func TestServeUntilSignal_ClosesConnectionsAfterTimeout(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	mux := http.NewServeMux()
	mux.HandleFunc("/stuck", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	quit := make(chan os.Signal, 1)
	go ServeUntilSignal(&http.Server{Handler: mux}, ln, quit, 50*time.Millisecond, nil)
	clientErr := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		_, err := client.Get("http://" + ln.Addr().String() + "/stuck")
		clientErr <- err
	}()
	<-started
	begin := time.Now()

	quit <- syscall.SIGTERM
	err = <-clientErr

	if err == nil || time.Since(begin) > time.Second {
		t.Fatalf("client err = %v after %v, want the connection closed soon after the 50ms shutdown timeout", err, time.Since(begin))
	}
}

func TestServeUntilSignal_ReturnsWhenServerShutsDownWithoutSignal(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.NewServeMux()}
	served := make(chan error, 1)
	go func() { served <- ServeUntilSignal(server, ln, make(chan os.Signal), time.Second, nil) }()
	waitForListener(t, ln.Addr().String())

	server.Shutdown(t.Context())

	select {
	case err := <-served:
		if err != nil {
			t.Fatalf("ServeUntilSignal = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ServeUntilSignal did not return after Shutdown without a signal")
	}
}

func TestServeUntilSignal_ReturnsServeError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()

	err = ServeUntilSignal(&http.Server{}, ln, make(chan os.Signal), time.Second, nil)

	if err == nil {
		t.Fatal("ServeUntilSignal = nil, want the Serve error from a closed listener")
	}
}

func waitForListener(t *testing.T, addr string) {
	t.Helper()
	for range 50 {
		if conn, err := net.Dial("tcp", addr); err == nil {
			conn.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("listener %s not ready", addr)
}
