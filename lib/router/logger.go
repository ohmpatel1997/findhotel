package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	zlog "github.com/ohmpatel1997/findhotel/lib/log"
)

const loggerKey = ctxKey("rlogger")

type ctxKey string

func LoggerAndRecover(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		sw := statusWriter{ResponseWriter: w}

		defer func(r *http.Request) {
			err, ok := recover().(error)
			if ok && err != nil {
				f := zlog.ParamsType{
					// err value from recover can be a non-error type
					"error":  fmt.Sprintf("%v", err),
					"host":   r.Host,
					"method": r.Method,
					"path":   r.URL.Path,
					"status": http.StatusInternalServerError,
				}

				zlog.Logger().Error("ROUTER ERROR", err, f)

				jsonBody, _ := json.Marshal(map[string]string{
					"error": "There was an internal server error",
				})

				w.WriteHeader(http.StatusInternalServerError)
				w.Write(jsonBody)
			}
		}(r)

		start := time.Now()

		next.ServeHTTP(&sw, r)

		duration := time.Now().Sub(start)

		zlog.Logger().Info("ACCESS", zlog.ParamsType{
			"host":           r.Host,
			"method":         r.Method,
			"path":           r.URL.Path,
			"status":         sw.status,
			"content_length": sw.length,
			"duration":       duration,
		})
	}

	return http.HandlerFunc(fn)
}
