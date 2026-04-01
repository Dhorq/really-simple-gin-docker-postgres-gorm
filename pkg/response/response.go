package response

import "github.com/gin-gonic/gin"

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Status:  200,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Response{
		Status:  status,
		Message: message,
		Data:    nil,
	})
}
