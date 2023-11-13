package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GoodsLoad(r *gin.Engine) *gin.Engine {
	g := r.Group("/goods")
	g.GET("/c", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"goods create": "success"})
	})
	return r
}
