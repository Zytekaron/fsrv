package server

import (
	"fsrv/src/config"
	"fsrv/src/database/dbutil"
	"fsrv/src/server/handlers"
	"fsrv/src/server/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	Dbi dbutil.DBInterface
	Cfg *config.Config
}

func New(DBInterface dbutil.DBInterface, config *config.Config) *Server {
	return &Server{DBInterface, config}
}

func (s *Server) Start(addr string) error {
	r := gin.Default()
	r.Use(middleware.UnifiedRateLimit(s.Dbi, s.Cfg.Server))
	r.Use(middleware.Auth(s.Dbi, s.Cfg.FileManager))

	handlers.New().Register(r)
	return http.ListenAndServe(addr, r)
}
