package order

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routers(e *gin.Engine) {
	e.GET("/order", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"order create": "success"})
	})
}
