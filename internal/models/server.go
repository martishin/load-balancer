package models

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Server struct {
	URL   *url.URL
	Alive bool
	mux   sync.RWMutex
	Proxy *httputil.ReverseProxy
}

func (s *Server) IsAlive() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.Alive
}

func (s *Server) SetAlive(status bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Alive = status
}
