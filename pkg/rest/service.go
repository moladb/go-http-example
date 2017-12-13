package rest

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
	API         string
	Path        string
	Method      string
	HandlerFunc gin.HandlerFunc
}

type Service interface {
	GetAPIGroup() string
	ListHandlers() []Handler
}
