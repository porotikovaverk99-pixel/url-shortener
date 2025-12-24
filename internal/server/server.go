package server

import (
	"net/http"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	router *chi.Mux
	addr string
}

func New(addr string) *Server {
	return &Server{
		router: chi.NewRouter(),
		addr: addr, 
	}
}

func (s *Server) RegisterHandler(pattern string, handler http.HandlerFunc) {
	s.router.HandleFunc(pattern, handler)
}

func (s *Server) RegisterHandle(pattern string, handler http.Handler) {
	s.router.Handle(pattern, handler)
}

func (s *Server) Post(pattern string, handler http.HandlerFunc) {
	s.router.Post(pattern, handler)
}

func (s *Server) Get(pattern string, handler http.HandlerFunc) {
	s.router.Get(pattern, handler)
}

func (s *Server) Put(pattern string, handler http.HandlerFunc) {
	s.router.Put(pattern, handler)
}

func (s *Server) Delete(pattern string, handler http.HandlerFunc) {
	s.router.Delete(pattern, handler)
}

func (s *Server) Run() error {
	return http.ListenAndServe(s.addr, s.router)
}