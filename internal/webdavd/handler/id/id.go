package id

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
)

// A Generator generates a request id from a request.
// Doesn't have to look at the request, can just generate
// a numeric sequence.
//
type Generator func(*http.Request) string

func Handle(generate Generator, next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		ctx = WithRequestId(ctx, generate(req))
		next.ServeHTTP(w, req.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

// id holds the next sequential request id.
//
var id uint64

// Generate is an example request-id generator that returns a
// numeric sequence starting with "0".
//
func Generate(req *http.Request) string {
	id := atomic.AddUint64(&id, 1)
	return strconv.FormatUint(id, 10)
}

const RequestIdKey = "request-id"

// WithRequestId returns a context with the request-id set to id.
//
func WithRequestId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIdKey, id)
}

// RequestId returns the context's request id, if it exists.
//
func RequestId(ctx context.Context) string {
	id, ok := ctx.Value(RequestIdKey).(string)
	if !ok {
		return ""
	}
	return id
}
