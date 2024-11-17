package http_utils

import (
	"log"
	"net/http"
	"time"
)

func ApplyMiddlewares(server http.Handler,
	handlers ...func(http.Handler) http.Handler) http.Handler {
	for _, h := range handlers {
		server = h(server)
	}
	return server
}

func StatusLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		lw := &loggingResponseWriter{w, http.StatusOK}
		next.ServeHTTP(lw, r)
		t2 := time.Now()
		log.Printf("[%s] %s %s %d %dms\n", r.Method, r.URL.Path, r.RemoteAddr,
			lw.statusCode, t2.Sub(t1).Milliseconds())
	})
}

// Custom response writer to intercept status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Flush() {
	flusher, ok := lrw.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

// func HttpLogging(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
// 		next.ServeHTTP(w, r)
// 	})
// }
