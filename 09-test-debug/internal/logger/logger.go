package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	output io.Writer
	level  LogLevel
	mu     sync.Mutex
}

type LogEntry struct {
	Time       time.Time              `json:"time"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	Duration   *float64               `json:"duration_ms,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	File       string                 `json:"file,omitempty"`
	Line       int                    `json:"line,omitempty"`
}

func NewLogger(output io.Writer, level LogLevel) *Logger {
	return &Logger{
		output: output,
		level:  level,
	}
}

func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}, traceID string, duration *float64, err error) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Time:     time.Now().UTC(),
		Level:    level.String(),
		Message:  message,
		Fields:   fields,
		TraceID:  traceID,
		Duration: duration,
	}

	// エラー情報の追加
	if err != nil {
		entry.Error = err.Error()
	}

	// 呼び出し元の情報を取得
	if pc, file, line, ok := runtime.Caller(2); ok {
		entry.File = fmt.Sprintf("%s:%d", file, line)
		entry.Line = line

		// 関数名も取得
		if fn := runtime.FuncForPC(pc); fn != nil {
			if fields == nil {
				fields = make(map[string]interface{})
			}
			fields["function"] = fn.Name()
			entry.Fields = fields
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	encoder := json.NewEncoder(l.output)
	encoder.Encode(entry)
}

func (l *Logger) Debug(message string, fields map[string]interface{}, traceID string) {
	l.log(DEBUG, message, fields, traceID, nil, nil)
}

func (l *Logger) Info(message string, fields map[string]interface{}, traceID string) {
	l.log(INFO, message, fields, traceID, nil, nil)
}

func (l *Logger) Warn(message string, fields map[string]interface{}, traceID string) {
	l.log(WARN, message, fields, traceID, nil, nil)
}

func (l *Logger) Error(message string, fields map[string]interface{}, traceID string) {
	l.log(ERROR, message, fields, traceID, nil, nil)
}

func (l *Logger) ErrorWithStack(message string, fields map[string]interface{}, traceID string, err error) {
	if level := ERROR; level >= l.level {
		entry := LogEntry{
			Time:       time.Now().UTC(),
			Level:      level.String(),
			Message:    message,
			Fields:     fields,
			TraceID:    traceID,
			StackTrace: getStackTrace(),
		}

		if err != nil {
			entry.Error = err.Error()
		}

		l.mu.Lock()
		defer l.mu.Unlock()

		encoder := json.NewEncoder(l.output)
		encoder.Encode(entry)
	}
}

func (l *Logger) InfoWithDuration(message string, fields map[string]interface{}, traceID string, duration *float64) {
	l.log(INFO, message, fields, traceID, duration, nil)
}

// Metrics collects application metrics
type Metrics struct {
	RequestCount   uint64
	ErrorCount     uint64
	TotalDuration  int64 // nanoseconds
	ActiveRequests int32
	mu             sync.RWMutex
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) RecordRequest(duration time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RequestCount++
	m.TotalDuration += int64(duration)

	if isError {
		m.ErrorCount++
	}
}

func (m *Metrics) IncrementActiveRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveRequests++
}

func (m *Metrics) DecrementActiveRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveRequests--
}

func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var avgDuration float64
	if m.RequestCount > 0 {
		avgDuration = float64(m.TotalDuration) / float64(m.RequestCount) / 1e6 // Convert to milliseconds
	}

	var errorRate float64
	if m.RequestCount > 0 {
		errorRate = float64(m.ErrorCount) / float64(m.RequestCount) * 100
	}

	return map[string]interface{}{
		"request_count":     m.RequestCount,
		"error_count":       m.ErrorCount,
		"error_rate":        fmt.Sprintf("%.2f%%", errorRate),
		"avg_duration_ms":   fmt.Sprintf("%.2f", avgDuration),
		"active_requests":   m.ActiveRequests,
		"total_duration_ms": float64(m.TotalDuration) / 1e6,
	}
}

func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RequestCount = 0
	m.ErrorCount = 0
	m.TotalDuration = 0
	m.ActiveRequests = 0
}

// TestMetrics provides metrics for testing purposes
type TestMetrics struct {
	TestsRun    int
	TestsPassed int
	TestsFailed int
	Coverage    float64
	Duration    time.Duration
}

func NewTestMetrics() *TestMetrics {
	return &TestMetrics{}
}

func (tm *TestMetrics) RecordTest(passed bool, duration time.Duration) {
	tm.TestsRun++
	tm.Duration += duration

	if passed {
		tm.TestsPassed++
	} else {
		tm.TestsFailed++
	}
}

func (tm *TestMetrics) GetSummary() map[string]interface{} {
	successRate := float64(0)
	if tm.TestsRun > 0 {
		successRate = float64(tm.TestsPassed) / float64(tm.TestsRun) * 100
	}

	return map[string]interface{}{
		"tests_run":     tm.TestsRun,
		"tests_passed":  tm.TestsPassed,
		"tests_failed":  tm.TestsFailed,
		"success_rate":  fmt.Sprintf("%.2f%%", successRate),
		"total_duration": tm.Duration.String(),
		"coverage":      fmt.Sprintf("%.2f%%", tm.Coverage),
	}
}

// ユーティリティ関数
func getStackTrace() string {
	buf := make([]byte, 1024*16)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// テスト用のメモリ出力ロガー
type MemoryLogger struct {
	entries []LogEntry
	mu      sync.RWMutex
}

func NewMemoryLogger() *MemoryLogger {
	return &MemoryLogger{
		entries: make([]LogEntry, 0),
	}
}

func (ml *MemoryLogger) Write(p []byte) (int, error) {
	var entry LogEntry
	if err := json.Unmarshal(p, &entry); err == nil {
		ml.mu.Lock()
		ml.entries = append(ml.entries, entry)
		ml.mu.Unlock()
	}
	return len(p), nil
}

func (ml *MemoryLogger) GetEntries() []LogEntry {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	// コピーを返す
	entries := make([]LogEntry, len(ml.entries))
	copy(entries, ml.entries)
	return entries
}

func (ml *MemoryLogger) Clear() {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.entries = ml.entries[:0]
}

func (ml *MemoryLogger) Count() int {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return len(ml.entries)
}

func (ml *MemoryLogger) GetLastEntry() *LogEntry {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	if len(ml.entries) == 0 {
		return nil
	}
	return &ml.entries[len(ml.entries)-1]
}