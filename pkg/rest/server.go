package rest

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moladb/ginprom"
	"github.com/moladb/go-http-example/pkg/version"
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

func (s *Server) RegisterServiceGroup(svc ServiceGroup) {
	// TODO: register these info into a discovery api
	apiGroup := svc.GetAPIGroup()
	apiGroup = "/" + strings.Trim(apiGroup, "/")
	group := s.router.Group(apiGroup)
	handlers := svc.ListHandlers()
	for _, h := range handlers {
		group.Handle(h.Method, "/"+strings.TrimLeft(h.Path, "/"),
			s.decorateHandler(apiGroup, h.Resource.Name, h.HandlerFunc))
		s.registry.AddGroupResource(apiGroup, h.Resource)
	}
}

func (s *Server) RegisterService(svc Service) {
	handlers := svc.ListHandlers()
	for _, h := range handlers {
		s.router.Handle(h.Method, "/"+strings.TrimLeft(h.Path, "/"),
			s.decorateHandler("", h.Resource.Name, h.HandlerFunc))
		s.registry.AddResource(h.Resource)
	}
}

func (s *Server) Run() error {
	s.router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK,
			gin.H{
				"version":    version.VERSION,
				"build_date": version.BUILDDATE,
				"go_version": version.GOVERSION,
			})
	})
	if s.config.EnablePProf {
		s.RegisterServiceGroup(newPProfService())
	}
	if s.config.EnableAPIMetrics {
		s.RegisterService(newMetricsService())
	}
	s.RegisterService(newDiscoveryService(s.registry))
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
