package server

import (
	"context"
	"fmt"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-cygnus/middlewares"
	"go-cygnus/utils/logging"
)

const (
	WebServerShutDown = 5 * time.Second
)

// WebServer throw an error when register already existing path
type WebServer struct {
	*gin.Engine
	logger        *logging.ConvenientErrorLogger
	urlGroupPaths []string
	underlying    *http.Server
}

// New create a default web server
func New() *WebServer {
	srv := &WebServer{
		Engine:        gin.New(),
		urlGroupPaths: make([]string, 4),
	}
	srv.logger = logging.GetLogger("webserver").WithField("instance", fmt.Sprintf("%p", srv))
	srv.Engine.Use(
		gzip.Gzip(gzip.DefaultCompression),
		middlewares.AccessLogger(),
		gin.Recovery(),
	   sentrygin.New(sentrygin.Options{Repanic: true}),
    )

	srv.Swagger()

	return srv
}

// NewRouterGroup registers a RouterGroup on a server instance
func (s *WebServer) NewRouterGroup(prefix string) *gin.RouterGroup {
	for _, p := range s.urlGroupPaths {
		if p == prefix {
			s.logger.Fatalf("duplicate prefix path %q", prefix)
		}
	}

	return s.Group(prefix)
}

// Swagger Generate swagger apis documents
func (s *WebServer) Swagger() {
	// use ginSwagger middleware to serve the API docs
	s.GET("/v1/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// Run RunOn start go routine bind port and listen to kill chan
func (s *WebServer) Run(addr string) {
	// TODO: impl finer Server configs
	srv := &http.Server{
		Addr:    addr,
		Handler: s.Engine,
	}
	s.underlying = srv

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				s.logger.Info("underlying closed")
			} else {
				s.logger.WithError(err).Error("underlying crash")
			}
		}
	}()
}

// Shutdown Gracefully shutdown server with a timeout,
// caller listen to ctx.Done know shutdown process finish (may success or fail)
func (s *WebServer) Shutdown() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), WebServerShutDown)

	go func() {
		defer cancel()
		if err := s.underlying.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("shutdown failed")
		}

		s.logger.Info("done with shutdown")
	}()

	return ctx
}
