package handlers

import (
	"github.com/gin-gonic/gin"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/", h.Get)
	r.POST("/", h.Create)
	r.PATCH("/", h.Update)
	r.DELETE("/", h.Delete)
}
