package requestlog

import (
	"net"
	"net/http"
	"time"

	ctxUtil "github.com/bartosz121/webhooks-api/cmd/api/util/ctx"
	"github.com/rs/zerolog"
)

type responseStats struct {
	w    http.ResponseWriter
	code int
}

func (r *responseStats) Header() http.Header {
	return r.w.Header()
}

func (r *responseStats) WriteHeader(statusCode int) {
	if r.code != 0 {
		return
	}

	r.w.WriteHeader(statusCode)
	r.code = statusCode
}

func (r *responseStats) Write(p []byte) (n int, err error) {
	if r.code == 0 {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.w.Write(p)
	return
}

type Handler struct {
	handler http.Handler
	l       *zerolog.Logger
}

func NewHandler(h http.HandlerFunc, l *zerolog.Logger) *Handler {
	return &Handler{
		handler: h,
		l:       l,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	logEntry := &logEntry{
		RequestId:     ctxUtil.RequestId(r.Context()),
		ReceivedTime:  start,
		RequestMethod: r.Method,
		RequestUrl:    r.URL.String(),
		UserAgent:     r.UserAgent(),
		RemoteIp:      ipFromHostPort(r.RemoteAddr),
	}

	h.l.Info().
		Str("request-id", logEntry.RequestId).
		Str("received-time", logEntry.ReceivedTime.String()).
		Str("method", logEntry.RequestMethod).
		Str("url", logEntry.RequestUrl).
		Str("user-agent", logEntry.UserAgent).
		Str("remote-ip", logEntry.RemoteIp).
		Msg("received")

	w2 := &responseStats{w: w}

	h.handler.ServeHTTP(w2, r)

	logEntry.Status = w2.code
	latency := time.Since(start)

	h.l.Info().
		Str("request-id", logEntry.RequestId).
		Str("received-time", logEntry.ReceivedTime.String()).
		Str("method", logEntry.RequestMethod).
		Str("url", logEntry.RequestUrl).
		Str("user-agent", logEntry.UserAgent).
		Str("remote-ip", logEntry.RemoteIp).
		Dur("latency", latency).
		Int("status", logEntry.Status).
		Msg("responded")
}

type logEntry struct {
	RequestId     string
	ReceivedTime  time.Time
	RequestMethod string
	RequestUrl    string
	UserAgent     string
	RemoteIp      string

	Status  int
	Latency time.Duration
}

func ipFromHostPort(hp string) string {
	h, _, err := net.SplitHostPort(hp)
	if err != nil {
		return ""
	}
	if len(h) > 0 && h[0] == '[' {
		return h[1 : len(h)-1]
	}
	return h
}
