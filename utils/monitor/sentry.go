package monitor

import (
	"net"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"

	"go-cygnus/models"
	"go-cygnus/utils"
	"go-cygnus/utils/logging"
)

var tokenList = []string{"X-User-Token"}

func SentryInit() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              utils.SysConfig.SentryConf.Dsn,
		Environment:      utils.SysConfig.SentryConf.Environment,
		AttachStacktrace: true,
	}); err != nil {
		logging.GetLogger("root").Infof("Sentry initialization failed: %s", err.Error())
	}
}

type Sentry struct {
	C   *gin.Context
	Err error
}

func (s *Sentry) QueryString() map[string]string {
	queryMap := make(map[string]string)
	for k, v := range s.C.Request.URL.Query() {
		queryMap[k] = strings.Join(v, ",")
	}

	return queryMap
}

func (s *Sentry) URL() string {
	proto := "http"
	if s.C.Request.TLS != nil || s.C.Request.Header.Get("X-Forwarded-Proto") == "https" {
		proto = "https"
	}

	return proto + "://" + s.C.Request.Host + s.C.Request.URL.Path
}

func (s *Sentry) Header() map[string]string {
	headersMap := make(map[string]string)
	for k, v := range s.C.Request.Header {
		headersMap[k] = strings.Join(v, ",")

		if funk.ContainsString(tokenList, k) {
			headersMap[k] = "****************************************************************"
		}
	}

	return headersMap
}

func (s *Sentry) Environment() map[string]string {
	environment := make(map[string]string)
	if addr, port, err := net.SplitHostPort(s.C.Request.RemoteAddr); err == nil {
		environment = map[string]string{"REMOTE_ADDR": addr, "REMOTE_PORT": port}
	}

	return environment
}

func (s *Sentry) User() (sentryUser sentry.User) {
	authUser, exist := s.C.Get("user")

	if exist {
		sentryUser.Username = authUser.(models.AuthUser).Username
		sentryUser.Email = authUser.(models.AuthUser).Email
	}

	return
}

func (s *Sentry) Post() {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetContext("Request", map[string]interface{}{
			"Method": s.C.Request.Method,
			"URL":    s.URL(),
		})
		scope.SetContext("Query Params", s.QueryString())
		scope.SetContext("Headers", s.Header())
		scope.SetContext("Environment", s.Environment())
		scope.SetUser(s.User())
		sentry.CaptureException(s.Err)
	})
}
