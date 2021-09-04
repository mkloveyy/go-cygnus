package apis

import (
	"fmt"
	"go-cygnus/dto"
	"net/http"
	"runtime"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"go-cygnus/middlewares"
	"go-cygnus/utils/logging"
	"go-cygnus/utils/server"
)

const (
	RuntimeCallerSkip = 2
)

// apis web server instance, submodule can register gin.RouterGroup on it
var (
	WebAPIServer = server.New()

	v1Router      = WebAPIServer.NewRouterGroup("v1")
	apiRootLogger = logging.GetLogger("apis")

	_ = v1Router.Use(middlewares.APINormalErrorHandler(apiRootLogger))
)

func init() {
	pprof.Register(WebAPIServer.Engine)

	v1Router.GET("health/check", HealthCheck)
}

// HealthCheck godoc
// @Summary health check
// @Description health check
// @Tags Health
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.BaseRsp
// @Router /health/check [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.BaseRsp{
		Message: "ok",
	})
}

// C is extension of gin.Context
type C struct {
	*gin.Context
}

func (c C) Logger() *logging.ConvenientErrorLogger {
	if l, ok := c.Get(middlewares.ContextKeyLogger); ok {
		return l.(*logging.ConvenientErrorLogger)
	}

	// same as apiRootLogger, as a fallback
	return apiRootLogger
}

// SetErr generate ApiError for ErrorAbort middleware, it should be direct called
// as it can locate error line number
func (c C) SetErr(err error, code ...int) {
	_, fileName, line, ok := runtime.Caller(RuntimeCallerSkip)
	if !ok {
		fileName = "???"
		line = 0
	}

	// TODO: use sentry to capture stack, insert into ApiError
	w := &middlewares.APIError{
		Origin:  err,
		ErrLine: fmt.Sprintf("%s:%d", fileName, line),
	}

	if len(code) > 0 {
		w.AsHTTPCode = code[0]
	}

	// send to sentry
	// if w.Code() >= http.StatusInternalServerError {
	// 	 sentry := &monitor.Sentry{
	//	 	C:   c.Context,
	//	 	Err: err,
	//	 }
	//	 sentry.Post()
	// }

	_ = c.Error(w)
}
