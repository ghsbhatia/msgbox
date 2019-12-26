package middleware

import (
	"net/http"
	"time"

	"github.com/ghsbhatia/msgbox/pkg/ctxlog"
	"github.com/go-kit/kit/log"
)

// HTTPInterceptor wraps an http.Handler and a log.Logger,
// and performs structured request logging.
type HTTPInterceptor struct {
	handler http.Handler
	logger  log.Logger
}

func NewHTTPInterceptor(handler http.Handler, logger log.Logger) *HTTPInterceptor {
	return &HTTPInterceptor{handler, logger}
}

// ServeHTTP implements http.Handler.
func (mw *HTTPInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		iw             = &interceptingWriter{http.StatusOK, w}
		ctx, ctxlogger = ctxlog.NewLogger(r.Context(), "http_method", r.Method, "http_path", r.URL.Path)
	)

	defer func(begin time.Time) {
		ctxlogger.Log("http_status_code", iw.code, "http_duration", time.Since(begin))
		mw.logger.Log(ctxlogger.Keyvals()...)
	}(time.Now())

	mw.handler.ServeHTTP(iw, r.WithContext(ctx))
}

type interceptingWriter struct {
	code int
	http.ResponseWriter
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}
