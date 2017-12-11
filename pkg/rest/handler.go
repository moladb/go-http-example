package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/moladb/go-http-example/pkg/version"
)

func (s *Server) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK,
		gin.H{
			"version":    version.VERSION,
			"build_time": version.BUILDDATE,
			"go_version": version.GOVERSION,
		})
}

func (s *Server) GetKVs(c *gin.Context) {
	key := c.Param("key")
	val, ok := s.getKV(key)
	if ok {
		c.JSON(http.StatusOK, gin.H{"value": val})
		return
	}
	c.Status(http.StatusNotFound)
}

func (s *Server) PutKV(c *gin.Context) {
	key := c.Param("key")
	var val struct {
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindWith(&val, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(val.Value) > maxDataLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exceed max_data_len(512K)"})
		return
	}
	s.putKV(key, val.Value)
	c.Status(http.StatusOK)
}

func (s *Server) DeleteKVs(c *gin.Context) {
	key := c.Param("key")
	s.deleteKV(key)
	c.Status(http.StatusOK)
}
