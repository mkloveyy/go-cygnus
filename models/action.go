package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	"go-cygnus/utils/db"
)

const (
	DEBUG   = 1
	INFO    = 2
	WARNING = 3
	ERROR   = 4

	InfoCode  = 300
	WarnCode  = 400
	ErrorCode = 500
)

type Client struct {
	IP string `json:"ip"`
}

type Host struct {
	Host string `json:"host"`
}

type Request struct {
	Path  string  `json:"path"`
	Data  db.JSON `json:"data,omitempty" sql:"type:json"`
	ReqID string  `json:"request_id"`
}

type Response struct {
	StatusCode int     `json:"status_code"`
	Data       db.JSON `json:"data,omitempty" sql:"type:json"`
}

type ResponseData struct {
	Message string `json:"message"`
}

type Action struct {
	BaseModel
	Client    db.JSON `json:"client" sql:"type:json"`
	Server    db.JSON `json:"server" sql:"type:json"`
	Request   db.JSON `json:"request" sql:"type:json"`
	Response  db.JSON `json:"response" sql:"type:json"`
	Operation string  `json:"operation"`
	Detail    string  `json:"detail"`
	Level     int     `gorm:"default:1" json:"level"`
	Tag       int     `json:"tag"`
	User      string  `gorm:"default:'anonymous'" json:"user"`
}

func (a *Action) getMethodNameThroughHandler(c *gin.Context) (methodName string) {
	handlerName := strings.ReplaceAll(c.HandlerName(), "git.ctripcorp.com/captain/captain-cd/apis.", "")

	switch handlerName {
	default:
		return handlerName
	}
}

// Before base function
func (a *Action) Before(c *gin.Context, reqData []byte) (err error) {
	// Get request
	var request Request
	request.Path = c.Request.URL.Path
	request.Data = reqData

	if reqID, ok := c.Get("req_id"); ok {
		request.ReqID = reqID.(string)
	}

	if a.Request, err = json.Marshal(&request); err != nil {
		return
	}

	// Get server
	var server Host
	server.Host = c.Request.Host

	if a.Server, err = json.Marshal(&server); err != nil {
		return
	}

	// Get client
	var client Client
	client.IP = c.ClientIP()

	if a.Client, err = json.Marshal(&client); err != nil {
		return
	}

	// Get user
	if user, exists := c.Get("user"); exists {
		a.User = fmt.Sprintf("%s %s", user.(AuthUser).DisplayName, user.(AuthUser).Email)
	}

	// Do by custom func if defined, otherwise ignore
	// judge for apis complex handler
	methodName := a.getMethodNameThroughHandler(c)
	a.Operation = methodName
	in := []reflect.Value{reflect.ValueOf(c), reflect.ValueOf(reqData)}

	method := reflect.ValueOf(a).MethodByName("Before" + methodName)
	if method.IsValid() {
		if res := method.Call(in)[0].Interface(); res != nil {
			err = res.(error)
			return
		}
	}

	return db.Engine.Create(&a).Error
}

func (a *Action) After(c *gin.Context, status int, body []byte) (err error) {
	// Get response
	var response Response
	response.StatusCode = status
	response.Data = body

	switch {
	case response.StatusCode < InfoCode:
		a.Level = INFO
	case response.StatusCode < WarnCode:
		a.Level = WARNING
	case response.StatusCode >= WarnCode:
		a.Level = ERROR
	}

	if a.Response, err = json.Marshal(&response); err != nil {
		return
	}

	// Do by custom func if defined, otherwise ignore
	in := []reflect.Value{reflect.ValueOf(c), reflect.ValueOf(body)}

	// judge for apis complex handler
	methodName := a.getMethodNameThroughHandler(c)

	method := reflect.ValueOf(a).MethodByName("After" + methodName)
	if method.IsValid() {
		if res := method.Call(in)[0].Interface(); res != nil {
			err = res.(error)
			return
		}
	}

	// append error msg
	if a.Level > INFO {
		var msg = struct {
			Message string `json:"message"`
		}{}

		if err = json.Unmarshal(body, &msg); err != nil {
			return
		}

		a.Detail += fmt.Sprintf(" error: %s", msg.Message)
	}

	if err = db.Engine.Save(&a).Error; err != nil {
		return
	}

	return
}
