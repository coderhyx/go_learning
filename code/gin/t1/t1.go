package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// gin的helloWorld
func main() {
	// 1.创建路由
	// 默认使用了2个中间件Logger(), Recovery()
	r := gin.Default()
	r.GET("/t1", func(c *gin.Context) {
		// 指定重定向的URL
		c.Request.URL.Path = "/t2"
		r.HandleContext(c)
	})
	r.GET("/t2", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"welcome to ": "t2"})
	})
	r.Run(":8000")
}
