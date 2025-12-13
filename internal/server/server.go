package server

import (
	"net/http"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	route *chi.Mux
	addr string
}

func New(handler http.HandlerFunc, addr string) *Server {
	r := chi.NewRouter()
	r.HandleFunc("/", handler)
	r.HandleFunc("/{id}", handler)
	return &Server{
		route: r,
		addr: addr, 
	}
}

func (s *Server) Run() error {
	return http.ListenAndServe(s.addr, s.route)
}