package lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type StatusCoder interface {
	StatusCode() int
}

type httpError struct {
	Status  int
	Message string
}

func newHTTPError(status int) httpError {
	return httpError{
		Status:  status,
		Message: http.StatusText(status),
	}
}

func (e httpError) Error() string {
	return fmt.Sprintf("%d %s", e.Status, e.Message)
}

func (e httpError) StatusCode() int {
	return e.Status
}

var ErrHTTPBadRequest = newHTTPError(http.StatusBadRequest)
var ErrHTTPUnauthorized = newHTTPError(http.StatusUnauthorized)
var ErrHTTPForbidden = newHTTPError(http.StatusForbidden)
var ErrHTTPNotFound = newHTTPError(http.StatusNotFound)
var ErrHTTPConflict = newHTTPError(http.StatusConflict)
var ErrHTTPInternal = newHTTPError(http.StatusInternalServerError)

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

type Router func(pattern string, handler http.HandlerFunc)

type mwResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *mwResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Middleware(recover bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := MakeRandomID("req", 8)

		ctx1 := WithLogValue(r.Context(), "req_id", reqID)
		r = r.WithContext(ctx1)

		w1 := &mwResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		t0 := time.Now()

		var err any
		if recover {
			err = safeCall(func() {
				next(w1, r)
			})
		} else {
			next(w1, r)
		}

		if err != nil {
			SendHTTPError(w, 0, fmt.Errorf("recover: %v", err))
		}

		t1 := time.Now()

		status := w1.statusCode

		DefaultLog(ctx1).Info("http", "remote", r.RemoteAddr, "method", r.Method, "url", r.URL.String(), "status", status, "time_ms", t1.Sub(t0).Milliseconds(), "err", err)
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
			SendHTTPError(w, 0, err)
		} else if data, ok := res.(HTTPResponse); ok {
			SendHTTPResponse(w, data)
		} else {
			SendHTTPResponse(w, HTTPResponse{
				Status: http.StatusOK,
				Data:   res,
			})
		}
	}
}

func SendHTTPError(w http.ResponseWriter, status int, err error) {
	data := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}

	if status == 0 {
		if he, ok := err.(StatusCoder); ok {
			status = he.StatusCode()
		} else {
			var he httpError
			if errors.As(err, &he) {
				status = he.StatusCode()
			} else {
				status = http.StatusInternalServerError
			}
		}
	}

	SendHTTPResponse(w, HTTPResponse{
		Status: status,
		Data:   data,
	})
}

func SendHTTPResponse(w http.ResponseWriter, res HTTPResponse) {
	data, err := json.Marshal(res.Data)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("output error: " + err.Error()))
		return
	}

	w.WriteHeader(res.Status)
	w.Write(data)
}
