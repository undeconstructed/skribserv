package lib

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func safeCall(f func()) any {
	var err any

	func() {
		defer func() {
			rec := recover()
			if rec != nil {
				err = rec
			}
		}()

		f()
	}()

	return err
}

type mwResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *mwResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Middleware(log *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := MakeRandomID("req", 8)

		log := log.With("req_id", reqID)

		ctx := PutLogger(r.Context(), log)

		r = r.WithContext(ctx)

		w1 := &mwResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		t0 := time.Now()

		err := safeCall(func() {
			next(w1, r)
		})

		t1 := time.Now()

		status := w1.statusCode

		log.Info("http", "remote", r.RemoteAddr, "method", r.Method, "url", r.URL.String(), "status", status, "time_ms", t1.Sub(t0).Milliseconds(), "err", err)
	}
}

type HTTPResponse struct {
	Status int
	Data   any
}

type APIFunc func(ctx context.Context, r *http.Request) any

func APIHandler(next APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := next(r.Context(), r)
		if err, ok := res.(error); ok {
			data := struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			}

			sendHTTPResponse(w, HTTPResponse{
				Status: http.StatusOK,
				Data:   data,
			})
		} else if data, ok := res.(HTTPResponse); ok {
			sendHTTPResponse(w, data)
		} else {
			sendHTTPResponse(w, HTTPResponse{
				Status: http.StatusOK,
				Data:   res,
			})
		}
	}
}

func sendHTTPResponse(w http.ResponseWriter, res HTTPResponse) {
	data, err := json.Marshal(res.Data)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("output error: " + err.Error()))
		return
	}

	w.WriteHeader(res.Status)
	w.Write(data)
}
