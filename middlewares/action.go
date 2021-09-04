package middlewares

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v9"

	"go-cygnus/models"
	"go-cygnus/utils/logging"
	"go-cygnus/utils/validators"
)

var middlewareLogger = logging.GetLogger("middleware")

// BodyLogWriter extracts gin.ResponseWriter and has a byte.buffer to copy response data.
type BodyLogWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

func (w BodyLogWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ActionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var action models.Action
		// read origin body bytes
		reqData, err := c.GetRawData()
		if err != nil {
			middlewareLogger.WithError(err).Error("GetRawData from gin context err")
		}

		if err := action.Before(c, reqData); err != nil {
			middlewareLogger.WithError(err).Error("before action failed")
		}

		c.Set("action", action)
		// write reqData back to request body
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqData))

		// enable to get response data after c.Next()
		blw := &BodyLogWriter{Body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		status := blw.Status()
		body := blw.Body.Bytes()

		// Captain uses C{c}.SetErr(err, ...) instead of return c.JSON directly, APINormalErrorHandler deals with
		// err and then return c.JSON. But ActionMiddleware is called before APINormalErrorHandler so it cannot achieve
		// err code and message. So we retrieve err here to ensure ActionMiddleware can achieve err here.
		for _, ginErr := range c.Errors {
			if wErr, ok := ginErr.Err.(*APIError); ok {
				// format binding error
				err, ok := wErr.Origin.(validator.ValidationErrors)

				message := wErr.Error()
				if ok {
					message = validators.ValidatorErrorFormatter(err)
				}

				response := map[string]interface{}{
					"message": message,
				}

				status = wErr.Code()
				body, _ = json.Marshal(response)
			}
		}

		if err := action.After(c, status, body); err != nil {
			middlewareLogger.WithError(err).Error("after action failed")
		}
	}
}
