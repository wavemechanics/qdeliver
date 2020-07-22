package basicauth

import (
	"net/http"
	"os"

	"github.com/wavemechanics/qdeliver/internal/webdavd/handler/rlog"
)

type Checker func(user, pass string) bool

func Handle(realm string, check Checker, next http.Handler) http.Handler {
	if realm == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		realm = hostname
	}

	fn := func(w http.ResponseWriter, req *http.Request) {
		if user, pass, ok := req.BasicAuth(); !ok {
			rlog.Log(req, "no auth")
			w.Header().Set("WWW-Authenticate", `Basic realm=`+realm+`"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else if !check(user, pass) {
			rlog.Log(req, "bad auth")
			w.Header().Set("WWW-Authenticate", `Basic realm=`+realm+`"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, req)
		}
	}

	return http.HandlerFunc(fn)
}
