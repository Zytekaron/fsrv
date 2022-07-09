package server

import (
	"fsrv/src/server/handlers"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
}

func New() *Server {
	return &Server{}
}

func (s *Server) Start(addr string) error {
	r := gin.Default()
	handlers.New().Register(r)
	return http.ListenAndServe(addr, r)
}
