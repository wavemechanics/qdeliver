package rlog

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/wavemechanics/deliver/internal/webdavd/handler/id"
)

func Handle(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, req *http.Request) {
		Log(req, "%s %s %s", req.RemoteAddr, req.Method, req.URL)
		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

func Log(req *http.Request, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg = strings.TrimSuffix(msg, "\n")

	for _, line := range strings.Split(msg, "\n") {
		log.Printf("req %s %s\n", id.RequestId(req.Context()), line)
	}
}
