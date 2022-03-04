package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response return response
type Response struct {
	ErrorCode    int         `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
	Data         interface{} `json:"data"`
}

func ResponseError(c *gin.Context, code int, m string) {
	c.AbortWithStatusJSON(code, Response{
		ErrorCode:    code,
		ErrorMessage: m,
	})
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, Response{
		Data: data,
	})
}

func ResponseResult(c *gin.Context, res Response) {
	c.AbortWithStatusJSON(http.StatusOK, res)
}
