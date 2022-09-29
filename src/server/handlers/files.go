package handlers

import (
	"fsrv/src/database"
	"fsrv/src/filemanager"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	database    database.DBInterface
	fileManager *filemanager.FileManager
}

func New(db database.DBInterface, fm *filemanager.FileManager) *Handler {
	return &Handler{
		database:    db,
		fileManager: fm,
	}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/", h.Get())
	r.POST("/", h.Create())
	r.PATCH("/", h.Update())
	r.DELETE("/", h.Delete())
}
