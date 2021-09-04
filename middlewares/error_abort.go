package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
	"go.elastic.co/apm/module/apmlogrus"
	"gopkg.in/go-playground/validator.v9"
	"gorm.io/gorm"

	"go-cygnus/constants"
	"go-cygnus/utils/logging"
	"go-cygnus/utils/validators"
)

const (
	ContextKeyLogger = "req_logger"
	ContextKeyReqID  = "req_id"
)

// mapping of known error to resp http code, used by SetErr
var errorCodeMap = map[reflect.Type]int{
	reflect.TypeOf(json.SyntaxError{}):           http.StatusBadRequest,
	reflect.TypeOf(json.UnmarshalTypeError{}):    http.StatusBadRequest,
	reflect.TypeOf(json.MarshalerError{}):        http.StatusBadRequest,
	reflect.TypeOf(json.UnsupportedValueError{}): http.StatusBadRequest,
	reflect.TypeOf(validator.ValidationErrors{}): http.StatusBadRequest,
}

type ErrJSONDto struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	ReqID   string `json:"req_id"`
}

// APIError TODO: add tracestack?
type APIError struct {
	AsHTTPCode int // optional code, 0 then use errorCodeMap
	Origin     error
	ErrLine    string
}

func (w *APIError) Error() string {
	return fmt.Sprintf("apis error cuz by: %s", w.Origin.Error())
}

// Code err <> http XXX
func (w *APIError) Code() int {
	if w.Origin == gorm.ErrRecordNotFound {
		return http.StatusNotFound
	}

	if w.AsHTTPCode != 0 {
		return w.AsHTTPCode
	}

	if code, ok := errorCodeMap[reflect.TypeOf(w.Origin)]; ok {
		return code
	}

	return http.StatusInternalServerError
}

// APINormalErrorHandler HandlerWrapError check non exception error, log them under apis and return 400/500
func APINormalErrorHandler(logger *logging.ConvenientErrorLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// inject logger
		if _, ok := c.Get(ContextKeyLogger); !ok {
			traceContextFields := apmlogrus.TraceContext(c.Request.Context())

			reqID := uuid.New()
			c.Set(ContextKeyReqID, reqID)

			uri := c.Request.URL.RequestURI()
			l := logger.WithFields(traceContextFields).WithField(ContextKeyReqID, reqID).WithField("uri", uri)

			c.Set(ContextKeyLogger, l)
		}

		c.Next()

		for _, ginErr := range c.Errors {
			if wErr, ok := ginErr.Err.(*APIError); ok {
				httpCode := wErr.Code()

				// format binding error
				err, ok := wErr.Origin.(validator.ValidationErrors)

				message := wErr.Error()
				if ok {
					message = validators.ValidatorErrorFormatter(err)
				}

				dto := ErrJSONDto{
					Message: message,
					Code:    httpCode,
					ReqID:   "???",
				}
				reqID, _ := c.Get(ContextKeyReqID)
				dto.ReqID = reqID.(string)

				injected, _ := c.Get(ContextKeyLogger)
				l := injected.(*logging.ConvenientErrorLogger).
					WithField("line", wErr.ErrLine).
					WithField("stack", fmt.Sprintf("%+v", wErr.Origin))

				msg := fmt.Sprintf("Http %d cuz by %s", httpCode, wErr.Origin)

				if httpCode > constants.HTTPCode499 {
					l.Error(msg)
				} else {
					l.Debug(msg)
				}

				// c.AbortWithStatusJSON(httpCode, dto)
				c.JSON(httpCode, dto)
			}
		}
	}
}
