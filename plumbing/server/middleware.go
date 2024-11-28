package server

import (
	"net/http"
)

type Middleware func(Handler) Handler

func Chain(h Handler, middleware ...Middleware) Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func (s *Server) ApplyMiddleware(h Handler) Handler {
	if len(s.middlewares) == 0 {
		return h
	}
	return Chain(h, s.middlewares...)
}

func (s *Server) Use(middleware ...Middleware) {
	s.middlewares = append(s.middlewares, middleware...)
}

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}
