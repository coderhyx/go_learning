package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PayLoad(r *gin.Engine) *gin.Engine {
	g := r.Group("/pay")
	g.GET("/c", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"pay create": "success"})
	})
	return r
}
