package rest

import (
	"context"
	"log"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moladb/go-http-example/pkg/version"
)

const maxDataLen int = 512 * 1024 // 512K

type Config struct {
	BindAddr    string
	EnablePProf bool
}

type Server struct {
	config  Config
	router  *gin.Engine
	httpSrv *http.Server
	kvs     map[string]string
	kvLock  sync.RWMutex
}

func NewServer(config Config) *Server {
	s := &Server{
		config: config,
		router: gin.Default(),
		kvs:    make(map[string]string),
	}
	s.installHandlers(config.EnablePProf)
	return s
}

func (s *Server) Run() error {
	s.httpSrv = &http.Server{
		Addr:    s.config.BindAddr,
		Handler: s.router,
	}
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}

func (s *Server) getKV(key string) (string, bool) {
	s.kvLock.RLock()
	defer s.kvLock.RUnlock()
	val, ok := s.kvs[key]
	if !ok {
		return "", false
	}
	return val, true
}

func (s *Server) putKV(key, value string) {
	s.kvLock.Lock()
	defer s.kvLock.Unlock()
	s.kvs[key] = value
}

func (s *Server) exist(key string) bool {
	s.kvLock.RLock()
	defer s.kvLock.RUnlock()
	_, ok := s.kvs[key]
	return ok
}

func (s *Server) deleteKV(key string) {
	s.kvLock.Lock()
	defer s.kvLock.Unlock()
	delete(s.kvs, key)
}

func pprofHandler(h http.HandlerFunc) gin.HandlerFunc {
	handler := http.HandlerFunc(h)
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func (s *Server) installHandlers(enableProfile bool) {
	s.router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK,
			gin.H{
				"version":    version.VERSION,
				"build_time": version.BUILDDATE,
				"go_version": version.GOVERSION,
			})
	})

	s.router.GET("/v1/kv/*key", func(c *gin.Context) {
		key := c.Param("key")
		val, ok := s.getKV(key)
		if ok {
			c.JSON(http.StatusOK, gin.H{"value": val})
		}
		c.Status(http.StatusNotFound)
	})

	s.router.PUT("/v1/kv/*key", func(c *gin.Context) {
		key := c.Param("key")
		var val struct {
			Value string `json:"value" binding:"required"`
		}
		if err := c.BindJSON(&val); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(val.Value) > maxDataLen {
			c.JSON(http.StatusBadRequest, gin.H{"error": "exceed max_data_len(512K)"})
			return
		}
		s.putKV(key, val.Value)
		c.Status(http.StatusOK)
	})

	s.router.DELETE("/v1/kv/*key", func(c *gin.Context) {
		key := c.Param("key")
		s.deleteKV(key)
		c.Status(http.StatusOK)
	})

	if enableProfile {
		s.router.GET("/debug/pprof", pprofHandler(pprof.Index))
		s.router.GET("/debug/pprof/profile", pprofHandler(pprof.Profile))
		s.router.GET("/debug/pprof/symbol", pprofHandler(pprof.Symbol))
		s.router.GET("/debug/pprof/cmdline", pprofHandler(pprof.Cmdline))
	}
}
