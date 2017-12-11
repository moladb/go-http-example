package rest

import (
	"context"
	"log"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moladb/ginprom"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	s.router.GET("/version", ginprom.WithMetrics("/version", s.GetVersion))

	s.router.GET("/v1/kv/*key", ginprom.WithMetrics("/v1/kv", s.GetKVs))

	s.router.PUT("/v1/kv/*key", ginprom.WithMetrics("/v1/kv", s.PutKV))

	s.router.DELETE("/v1/kv/*key", ginprom.WithMetrics("/v1/kv", s.DeleteKVs))

	if enableProfile {
		s.router.GET("/debug/pprof", pprofHandler(pprof.Index))
		s.router.GET("/debug/pprof/profile", pprofHandler(pprof.Profile))
		s.router.GET("/debug/pprof/symbol", pprofHandler(pprof.Symbol))
		s.router.GET("/debug/pprof/cmdline", pprofHandler(pprof.Cmdline))
	}

	s.router.GET("/metrics", func() gin.HandlerFunc {
		handler := promhttp.Handler()
		return func(c *gin.Context) {
			handler.ServeHTTP(c.Writer, c.Request)
		}
	}())
}
