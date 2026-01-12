package response

import (
	"github.com/gin-gonic/gin"
)

func OK[T any](c *gin.Context, status int, data T) {
	c.JSON(status, SuccessResponse[T]{
		Success: true,
		Data:    data,
	})
}

func Fail(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}
