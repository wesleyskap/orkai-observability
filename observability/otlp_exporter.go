package observability

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type exportedSpan struct {
	span     Span
	duration time.Duration
}

type OTLPExporter struct {
	client      *http.Client
	endpoint    string
	serviceName string
	headers     map[string]string
	interval    time.Duration
	spansChan   chan exportedSpan
	logsChan    chan string
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewOTLPExporter(cfg Config) *OTLPExporter {
	if cfg.OTLPEndpoint == "" {
		return nil
	}
	d := cfg.ExportInterval
	if d <= 0 {
		d = 5 * time.Second
	}
	exp := &OTLPExporter{
		client:      &http.Client{Timeout: 10 * time.Second},
		endpoint:    cfg.OTLPEndpoint,
		serviceName: cfg.ServiceName,
		headers:     cfg.OTLPHeaders,
		interval:    d,
		spansChan:   make(chan exportedSpan, 4096),
		logsChan:    make(chan string, 4096),
		stopChan:    make(chan struct{}),
	}
	exp.wg.Add(1)
	go exp.worker()
	return exp
}

func (e *OTLPExporter) ExportSpan(span Span, duration time.Duration) {
	select {
	case e.spansChan <- exportedSpan{span: span, duration: duration}:
	default:
		Counter("observability_dropped_logs_total")
	}
}

func (e *OTLPExporter) ExportLog(logStr string) {
	select {
	case e.logsChan <- logStr:
	default:
		Counter("observability_dropped_logs_total")
	}
}

func (e *OTLPExporter) Close() {
	close(e.stopChan)
	e.wg.Wait()
}

func (e *OTLPExporter) worker() {
	defer e.wg.Done()
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	var spans []exportedSpan
	var logs []string
	for {
		select {
		case s := <-e.spansChan:
			spans = append(spans, s)
		case l := <-e.logsChan:
			logs = append(logs, l)
		case <-ticker.C:
			e.flush(spans, logs)
			spans = nil
			logs = nil
		case <-e.stopChan:
			e.drain(spans, logs)
			return
		}
	}
}

func (e *OTLPExporter) drain(spans []exportedSpan, logs []string) {
	for {
		select {
		case s := <-e.spansChan:
			spans = append(spans, s)
		default:
			goto drainLogs
		}
	}
drainLogs:
	for {
		select {
		case l := <-e.logsChan:
			logs = append(logs, l)
		default:
			goto send
		}
	}
send:
	e.flush(spans, logs)
}

func (e *OTLPExporter) flush(spans []exportedSpan, logs []string) {
	if len(spans) > 0 {
		e.sendTraces(spans)
	}
	if len(logs) > 0 {
		e.sendLogs(logs)
	}
}

func (e *OTLPExporter) sendTraces(spans []exportedSpan) {
	payload := e.buildTracesPayload(spans)
	body, err := json.Marshal(payload)
	if err != nil {
		Counter("observability_internal_errors_total")
		return
	}
	e.post(e.endpoint+"/v1/traces", body)
}

func (e *OTLPExporter) sendLogs(logs []string) {
	payload := e.buildLogsPayload(logs)
	body, err := json.Marshal(payload)
	if err != nil {
		Counter("observability_internal_errors_total")
		return
	}
	e.post(e.endpoint+"/v1/logs", body)
}

func (e *OTLPExporter) post(url string, body []byte) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		Counter("observability_internal_errors_total")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range e.headers {
		req.Header.Set(k, v)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		Counter("observability_internal_errors_total")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		Counter("observability_internal_errors_total")
		_, _ = io.Copy(io.Discard, resp.Body)
	}
}

type otlpAnyValue struct {
	StringValue string `json:"stringValue,omitempty"`
}

type otlpKeyValue struct {
	Key   string       `json:"key"`
	Value otlpAnyValue `json:"value"`
}

type otlpResource struct {
	Attributes []otlpKeyValue `json:"attributes"`
}

type otlpSpan struct {
	TraceID           string         `json:"traceId"`
	SpanID            string         `json:"spanId"`
	Name              string         `json:"name"`
	Kind              int            `json:"kind"`
	StartTimeUnixNano string         `json:"startTimeUnixNano"`
	EndTimeUnixNano   string         `json:"endTimeUnixNano"`
	Attributes        []otlpKeyValue `json:"attributes,omitempty"`
}

type otlpScopeSpan struct {
	Spans []otlpSpan `json:"spans"`
}

type otlpResourceSpan struct {
	Resource   otlpResource    `json:"resource"`
	ScopeSpans []otlpScopeSpan `json:"scopeSpans"`
}

type otlpTracesPayload struct {
	ResourceSpans []otlpResourceSpan `json:"resourceSpans"`
}

func (e *OTLPExporter) buildTracesPayload(spans []exportedSpan) otlpTracesPayload {
	var otSpans []otlpSpan
	for _, s := range spans {
		start := s.span.StartTime.UnixNano()
		end := start + s.duration.Nanoseconds()
		otSpans = append(otSpans, otlpSpan{
			TraceID:           s.span.TraceID,
			SpanID:            genSpanID(),
			Name:              s.span.Name,
			Kind:              1, // Internal
			StartTimeUnixNano: strconv.FormatInt(start, 10),
			EndTimeUnixNano:   strconv.FormatInt(end, 10),
		})
	}
	return otlpTracesPayload{
		ResourceSpans: []otlpResourceSpan{
			{
				Resource: otlpResource{
					Attributes: []otlpKeyValue{
						{Key: "service.name", Value: otlpAnyValue{StringValue: e.serviceName}},
					},
				},
				ScopeSpans: []otlpScopeSpan{{Spans: otSpans}},
			},
		},
	}
}

type otlpLogRecord struct {
	TimeUnixNano string         `json:"timeUnixNano"`
	Body         otlpAnyValue   `json:"body"`
	Attributes   []otlpKeyValue `json:"attributes,omitempty"`
	TraceID      string         `json:"traceId,omitempty"`
}

type otlpScopeLog struct {
	LogRecords []otlpLogRecord `json:"logRecords"`
}

type otlpResourceLog struct {
	Resource  otlpResource   `json:"resource"`
	ScopeLogs []otlpScopeLog `json:"scopeLogs"`
}

type otlpLogsPayload struct {
	ResourceLogs []otlpResourceLog `json:"resourceLogs"`
}

func (e *OTLPExporter) buildLogsPayload(logs []string) otlpLogsPayload {
	var records []otlpLogRecord
	for _, logStr := range logs {
		var m map[string]interface{}
		var bodyVal string
		traceID := ""
		if err := json.Unmarshal([]byte(logStr), &m); err == nil {
			if msg, ok := m["msg"].(string); ok {
				bodyVal = msg
			}
			if tid, ok := m["trace_id"].(string); ok {
				traceID = tid
			}
		} else {
			bodyVal = logStr
		}
		records = append(records, otlpLogRecord{
			TimeUnixNano: strconv.FormatInt(time.Now().UnixNano(), 10),
			Body:         otlpAnyValue{StringValue: bodyVal},
			TraceID:      traceID,
		})
	}
	return otlpLogsPayload{
		ResourceLogs: []otlpResourceLog{
			{
				Resource: otlpResource{
					Attributes: []otlpKeyValue{
						{Key: "service.name", Value: otlpAnyValue{StringValue: e.serviceName}},
					},
				},
				ScopeLogs: []otlpScopeLog{{LogRecords: records}},
			},
		},
	}
}

func genSpanID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
