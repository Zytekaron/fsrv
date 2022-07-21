package server

import (
	"fsrv/src/database"
	"fsrv/src/server/handlers"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	Database database.DBInterface
}

func New(dbInterface database.DBInterface) *Server {
	return &Server{dbInterface}
}

func (s *Server) Start(addr string) error {
	r := gin.Default()
	handlers.New().Register(r)
	return http.ListenAndServe(addr, r)
}
