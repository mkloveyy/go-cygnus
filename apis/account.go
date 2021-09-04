package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-cygnus/dto"
	"go-cygnus/middlewares"
)

func init() {
	account := v1Router.Group("accounts")
	{
		account.GET("", middlewares.PaginationMiddleware(), ListAccount)
		account.POST("", middlewares.ActionMiddleware(), AddAccount)
	}
}

// ListAccount godoc
// @Summary List account
// @Description get account list
// @Tags Account
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.ListAccountReq
// @Failure 400 {object} middlewares.ErrJSONDto
// @Failure 500 {object} middlewares.ErrJSONDto
// @Router /accounts [get]
func ListAccount(c *gin.Context) {
	s := dto.ListAccountReq{}
	if err := c.ShouldBindQuery(&s); err != nil {
		C{c}.SetErr(err, http.StatusBadRequest)
		return
	}

	rsp, err := s.List(c.MustGet("pagination").(dto.Pagination))
	if err != nil {
		C{c}.SetErr(err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &rsp)
}

// AddAccount godoc
// @Summary Add an account
// @Description adds an account
// @Tags Account
// @Accept  json
// @Produce  json
// @Param data body dto.AddAccountReq true "data"
// @Success 200 {object} dto.AddAccountRsp
// @Failure 400 {object} middlewares.ErrJSONDto
// @Failure 500 {object} middlewares.ErrJSONDto
// @Router /accounts [post]
func AddAccount(c *gin.Context) {
	s := dto.AddAccountReq{}
	if err := c.ShouldBindJSON(&s); err != nil {
		C{c}.SetErr(err, http.StatusBadRequest)
		return
	}

	rsp, err := s.Add()
	if err != nil {
		C{c}.SetErr(err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &rsp)
}
