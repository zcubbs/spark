package logger

import (
	"context"
	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

const (
	receivedRequestMsg = "received request"
)

func GrpcLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	startTime := time.Now()
	resp, err = handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	attributes := []interface{}{
		"protocol", "grpc",
		"method", info.FullMethod,
		"status", statusCode,
		"status_code", int(statusCode),
		"duration", duration,
	}
	if err != nil {
		attributes = append(attributes, "error", err)
	}

	log.Debug(receivedRequestMsg,
		attributes...,
	)

	return resp, err
}

type ResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	Body       []byte
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	r.Body = b
	return r.ResponseWriter.Write(b)
}

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		recorder := &ResponseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		handler.ServeHTTP(recorder, r)
		duration := time.Since(startTime)

		attributes := []interface{}{
			"protocol", "http",
			"method", r.Method,
			"path", r.RequestURI,
			"status", http.StatusText(recorder.statusCode),
			"status_code", recorder.statusCode,
			"duration", duration,
		}

		if recorder.statusCode >= 400 {
			attributes = append(attributes, "body", string(recorder.Body))
			log.Error(receivedRequestMsg,
				attributes...,
			)
		} else {
			log.Debug(receivedRequestMsg,
				attributes...,
			)
		}
	})
}
