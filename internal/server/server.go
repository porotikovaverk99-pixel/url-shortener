package server

import (
	"net/http"
)

type Server struct {
	route *http.ServeMux
	addr string
}

func New(handler *http.HandlerFunc) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	return &Server{
		route: mux,
		addr: ":8080",
	}
}

func (s *Server) Run() error {
	return http.ListenAndServe(s.addr, s.route)
}