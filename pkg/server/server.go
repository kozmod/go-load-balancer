package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

// Server holds the data about a server
type Server struct {
	URL          *url.URL
	alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func NewAlive(url *url.URL, lbFn func(w http.ResponseWriter, r *http.Request)) *Server {
	srv := &Server{URL: url, alive: true}
	proxy := NewProxy(srv, lbFn)
	srv.ReverseProxy = proxy
	return srv
}

// SetAlive for this backend
func (s *Server) SetAlive(alive bool) {
	s.mux.Lock()
	s.alive = alive
	s.mux.Unlock()
}

// IsAlive returns true when backend is alive
func (s *Server) IsAlive() (alive bool) {
	s.mux.RLock()
	alive = s.alive
	s.mux.RUnlock()
	return
}

//ServeHTTP uses ReverseProxy
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.ReverseProxy.ServeHTTP(rw, req)
}
