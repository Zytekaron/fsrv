package files

import (
	"fsrv/src/config"
	"fsrv/src/database"
	"fsrv/src/filemanager"
	"fsrv/src/server/admin/handlers"
	"fsrv/src/server/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	config      *config.Config
	database    database.DBInterface
	fileManager *filemanager.FileManager
}

func New(cfg *config.Config, db database.DBInterface, fm *filemanager.FileManager) *Server {
	return &Server{
		config:      cfg,
		database:    db,
		fileManager: fm,
	}
}

func (s *Server) Start(addr string) error {
	r := gin.Default()
	r.Use(middleware.GetIP())
	//r.Use(filesmw.UnifiedRateLimit(s.database, s.config.Server))
	//r.Use(filesmw.Auth(s.database, s.config.FileManager))
	//
	handlers.New(s.database, s.fileManager).Register(r)
	return http.ListenAndServe(addr, r)
}
