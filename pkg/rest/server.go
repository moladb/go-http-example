package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moladb/ginprom"
	"github.com/moladb/go-http-example/pkg/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	config          Config
	router          *gin.Engine
	httpSrv         *http.Server
	decorateHandler func(apiGroup, api string, h gin.HandlerFunc) gin.HandlerFunc
	registry        *serviceRegistry
}

func NewServer(config Config) *Server {
	s := &Server{
		config:   config,
		router:   gin.Default(),
		registry: newServiceRegistry(),
	}
	if config.EnableAPIMetrics {
		s.decorateHandler = func(apiGroup, resource string, h gin.HandlerFunc) gin.HandlerFunc {
			path := "/" + strings.Trim(apiGroup, "/") + "/" + strings.Trim(resource, "/")
			return ginprom.WithMetrics(path, h)
		}
	} else {
		s.decorateHandler = func(_, _ string, h gin.HandlerFunc) gin.HandlerFunc {
			return h
		}
	}
	return s
}

func (s *Server) RegisterService(svc Service) {
	// TODO: register these info into a discovery api
	apiGroup := svc.GetAPIGroup()
	apiGroup = "/" + strings.Trim(apiGroup, "/")
	group := s.router.Group(apiGroup)
	handlers := svc.ListHandlers()
	for _, h := range handlers {
		group.Handle(h.Method, "/"+strings.Trim(h.Path, "/"),
			s.decorateHandler(apiGroup, h.Resource.Name, h.HandlerFunc))
		s.registry.AddResource(apiGroup, h.Resource)
	}
}

func (s *Server) Run() error {
	// TODO: register discover api
	s.router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK,
			gin.H{
				"version":    version.VERSION,
				"build_date": version.BUILDDATE,
				"go_version": version.GOVERSION,
			})
	})
	s.router.GET("/apis", func(c *gin.Context) {
		c.JSON(http.StatusOK,
			gin.H{
				"apis": s.registry.ListAPIGroups(),
			})
	})
	s.router.GET("/apis/:apigroup", func(c *gin.Context) {
		var apiGroup string
		apiGroup = c.Param("apigroup")
		rs, ok := s.registry.ListResources(apiGroup)
		if !ok {
			c.JSON(http.StatusNotFound,
				gin.H{
					"error": fmt.Sprintf("APIGroup:%s not found", apiGroup),
				})
			return
		}
		c.JSON(http.StatusOK, rs)
	})
	if s.config.EnablePProf {
		s.router.GET("/debug/pprof", ginHandlerFunc(pprof.Index))
		s.router.GET("/debug/pprof/profile", ginHandlerFunc(pprof.Profile))
		s.router.GET("/debug/pprof/profile", ginHandlerFunc(pprof.Cmdline))
	}
	if s.config.EnableAPIMetrics {
		s.router.GET("/metrics", func() gin.HandlerFunc {
			h := promhttp.Handler()
			return func(c *gin.Context) {
				h.ServeHTTP(c.Writer, c.Request)
			}
		}())
	}
	s.httpSrv = &http.Server{
		Addr:    s.config.BindAddr,
		Handler: s.router,
	}
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(s.config.GraceShutdownTimeoutS)*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		log.Fatal("Server GraceShutdown:", err)
	}
}
