package middleware

import (
	"context"
	"net/http"

	"github.com/rs/xid"
)

type (
	requestIDType string
)

const (
	requestIDHeader requestIDType = "X-Request-ID"
)

func RequestID(ctx context.Context, req *http.Request) context.Context {
	reqID := req.Header.Get(string(requestIDHeader))
	if reqID == "" {
		reqID = xid.New().String()
	}
	return context.WithValue(ctx, requestIDHeader, reqID)
}

func SetRequestID(ctx context.Context, w http.ResponseWriter) context.Context {
	reqID := ctx.Value(requestIDHeader).(string)
	w.Header().Set(string(requestIDHeader), reqID)

	return ctx
}
