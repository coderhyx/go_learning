package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routers(e *gin.Engine) {
	r := e.Group("inventory")
	r.GET("create", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"pay create": "success"})
	})
}
