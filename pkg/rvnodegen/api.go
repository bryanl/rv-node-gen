package rvnodegen

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/bryanl/rv-node-gen/internal/log"
	"github.com/bryanl/rv-node-gen/internal/util"
)

const (
	contextKeyRequestID util.ContextKey = "requestID"
)

// API is the node gen api
type API struct {
	lister Lister
}

// NewAPI creates an instance of API.
func NewAPI(lister Lister) *API {
	a := &API{
		lister: lister,
	}
	return a
}

// Handler create a HTTP handler.
func (a *API) Handler(ctx context.Context) *mux.Router {
	logger := log.From(ctx)

	r := mux.NewRouter()
	r.Use(requestIDMiddleware)
	r.Use(logMiddleware(logger))
	r.Use(configureCORS)

	r.Handle("/v1/nodes", NewNodeHandler(a.lister)).Methods(http.MethodGet)
	r.Handle("/v1/ws", NewWebsocketHandler(a.lister))

	return r
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := uuid.New()
		ctx = context.WithValue(ctx, contextKeyRequestID, id.String())
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(logger *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := r.Context().Value(contextKeyRequestID).(string)

			wrapped := wrapResponseWriter(w)

			defer func() {
				logger.Info("HTTP",
					"method", r.Method,
					"status", wrapped.status,
					"uri", r.RequestURI,
					"latency", time.Since(start),
					"request_id", requestID,
				)
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

var _ http.Hijacker = &responseWriter{}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true

	return
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijacking is not supported")
	}

	return h.Hijack()
}
