package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-cygnus/dto"
)

const GinContextKeyPagination = "pagination"

type Pagination struct {
	Is       bool
	Page     int `form:"page"`
	PageSize int `form:"page_size" binding:"required_with=Page"`
	Offset   int
	Limit    int
}

func PaginationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var errMsg string

		var pagination = dto.Pagination{Is: true}

		// Check variable
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil {
			errMsg = "'page' must be integer"
		}

		pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if err != nil {
			errMsg = "'page_size' must be integer"
		}

		isPagination, err := strconv.ParseBool(c.DefaultQuery("is_pagination", "1"))
		if err != nil {
			errMsg = "'is_pagination' must be one of 0/1/true/false"
		}

		// Abort if query param is invalid
		if errMsg != "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid query param: " + errMsg,
			})
			c.Abort()
			return
		}

		// Set pagination value
		pagination.Page = page
		pagination.PageSize = pageSize

		switch {
		case !isPagination:
			// No pagination
			pagination.Is = false
			pagination.Limit = -1
			pagination.Page = -1
			pagination.PageSize = -1
		case page > 0 && pageSize > 0 && pageSize <= dto.MaxPageSize:
			// Pagination with input value
			pagination.Offset = (page - 1) * pageSize
			pagination.Limit = pageSize
		case page > 0 && pageSize > dto.MaxPageSize:
			// PageSize overload
			pagination.Limit = 200
		default:
			// Pagination with default value
			pagination.Limit = 20
		}

		c.Set(GinContextKeyPagination, pagination)

		c.Next()
	}
}
