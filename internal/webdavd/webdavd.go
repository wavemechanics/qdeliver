package webdavd

import (
	"crypto/subtle"
	"net/http"
	"net/http/httptest"

	"golang.org/x/net/webdav"

	"github.com/wavemechanics/deliver/internal/webdavd/handler/basicauth"
	"github.com/wavemechanics/deliver/internal/webdavd/handler/id"
	"github.com/wavemechanics/deliver/internal/webdavd/handler/rlog"
)

type Server struct {
	Dir  string // directory to serve
	Addr string // port server is listening on
	User string // login user
	Pass string // login password
}

func (s *Server) Start() func() {

	logger := func(r *http.Request, err error) {
		if err != nil {
			rlog.Log(r, "%+v", err)
		}
	}

	srv := &webdav.Handler{
		FileSystem: webdav.Dir(s.Dir),
		LockSystem: webdav.NewMemLS(),
		Logger:     logger,
	}

	checkpw := func(u, p string) bool {
		uok := subtle.ConstantTimeCompare([]byte(u), []byte(s.User)) == 1
		pok := subtle.ConstantTimeCompare([]byte(p), []byte(s.Pass)) == 1
		return uok && pok
	}

	mux := http.NewServeMux()
	mux.Handle("/", id.Handle(id.Generate, rlog.Handle(basicauth.Handle("", checkpw, srv))))

	server := httptest.NewServer(mux)
	s.Addr = server.URL

	return func() {
		server.Close()
	}
}
